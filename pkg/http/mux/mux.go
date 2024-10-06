package mux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypes "github.com/Motmedel/utils_go/pkg/http/mux/types"
	muxUtils "github.com/Motmedel/utils_go/pkg/http/mux/utils"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/accept_encoding"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	"io"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

var DefaultHeaders = map[string]string{
	"Cache-Control":                "no-store",
	"Cross-Origin-Opener-Policy":   "same-origin",
	"Cross-Origin-Embedder-Policy": "require-corp",
	"Cross-Origin-Resource-Policy": "same-origin",
	"Content-Security-Policy":      "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'",
	"X-Content-Type-Options":       "nosniff",
	"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
}

type parsedRequestBodyContextType struct{}

var ParsedRequestBodyContextKey parsedRequestBodyContextType

func WriteResponse(responseInfo *muxTypes.ResponseInfo, responseWriter http.ResponseWriter) error {
	if responseInfo == nil {
		return nil
	}

	if reflect.ValueOf(responseWriter).IsNil() {
		// TODO: ErrNilResponseWriter?
		return nil
	}

	var defaultHeaders map[string]string
	muxResponseWriter, ok := responseWriter.(*muxTypes.ResponseWriter)
	if ok {
		defaultHeaders = muxResponseWriter.DefaultHeaders
	} else {
		defaultHeaders = make(map[string]string)
	}
	skippedDefaultHeadersSet := make(map[string]struct{})

	responseWriterHeader := responseWriter.Header()
	for _, header := range responseInfo.Headers {
		if header == nil {
			continue
		}

		if strings.ToLower(header.Name) == "content-type" && len(responseInfo.Body) == 0 {
			continue
		}

		if _, ok := defaultHeaders[header.Name]; ok {
			if header.Overwrite {
				skippedDefaultHeadersSet[header.Name] = struct{}{}
			} else {
				continue
			}
		}

		responseWriterHeader.Set(header.Name, header.Value)
	}
	for headerName, headerValue := range defaultHeaders {
		if _, ok := skippedDefaultHeadersSet[headerName]; ok {
			continue
		}
		responseWriterHeader.Set(headerName, headerValue)
	}

	if responseInfo.StatusCode != 0 {
		responseWriter.WriteHeader(responseInfo.StatusCode)
	}

	if _, err := responseWriter.Write(responseInfo.Body); err != nil {
		return &motmedelErrors.CauseError{
			Message: "An error occurred when writing a response body.",
			Cause:   err,
		}
	}

	return nil
}

func PerformErrorResponse(
	responseWriter http.ResponseWriter,
	request *http.Request,
	problemDetail *problem_detail.ProblemDetail,
	responseHeaders []*muxTypes.HeaderEntry,
) {
	logger := motmedelLog.GetLoggerFromCtxWithDefault(request.Context(), nil)

	if problemDetail == nil {
		motmedelLog.LogError(
			"The error response problem detail is nil.",
			muxErrors.ErrNilErrorResponseProblemDetail,
			logger,
		)
		problemDetail = problem_detail.MakeInternalServerErrorProblemDetail("", nil)
		responseHeaders = nil
	}

	if problemDetail.Status == 0 {
		motmedelLog.LogError(
			"The error response problem detail status is unset.",
			muxErrors.ErrUnsetErrorResponseProblemDetailStatus,
			logger,
		)
		problemDetail = problem_detail.MakeInternalServerErrorProblemDetail("", nil)
		responseHeaders = nil
	}

	statusCode := problemDetail.Status
	responseBody, err := problemDetail.Bytes()
	if err != nil {
		motmedelLog.LogError("An error occurred when converting a problem detail into bytes.", err, logger)

		statusCode = http.StatusInternalServerError
		responseHeaders = nil
	}

	if responseHeaders == nil {
		responseHeaders = []*muxTypes.HeaderEntry{{Name: "Content-Type", Value: "application/problem+json"}}
	} else {
		for i, header := range responseHeaders {
			if strings.ToLower(header.Name) == "content-type" {
				responseHeaders[i] = nil
			}
		}
		responseHeaders = append(
			responseHeaders,
			&muxTypes.HeaderEntry{Name: "Content-Type", Value: "application/problem+json"},
		)
	}

	responseInfo := &muxTypes.ResponseInfo{StatusCode: statusCode, Body: responseBody, Headers: responseHeaders}

	if err := WriteResponse(responseInfo, responseWriter); err != nil {
		motmedelLog.LogError("An error occurred when writing a response.", err, logger)
	}
}

func DefaultClientErrorHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	requestBody []byte,
	problemDetail *problem_detail.ProblemDetail,
	headers []*muxTypes.HeaderEntry,
	err error,
) {
	if err != nil {
		motmedelLog.LogWarning(
			"A client error occurred.",
			err,
			motmedelLog.GetLoggerFromCtxWithDefault(request.Context(), nil),
		)
	}
	if problemDetail == nil {
		problemDetail = problem_detail.MakeBadRequestProblemDetail("", nil)
	}
	PerformErrorResponse(responseWriter, request, problemDetail, headers)
}

func DefaultServerErrorHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	requestBody []byte,
	problemDetail *problem_detail.ProblemDetail,
	responseHeaders []*muxTypes.HeaderEntry,
	err error,
) {
	if err != nil {
		motmedelLog.LogError(
			"A server error occurred.",
			err,
			motmedelLog.GetLoggerFromCtxWithDefault(request.Context(), nil),
		)
	}
	if problemDetail == nil {
		problemDetail = problem_detail.MakeInternalServerErrorProblemDetail("", nil)
	}
	PerformErrorResponse(responseWriter, request, problemDetail, responseHeaders)
}

type Mux struct {
	HandlerSpecificationMap map[string]map[string]*muxTypes.HandlerSpecification
	SetContextKeyValuePairs [][2]any
	ClientErrorHandler      func(
		http.ResponseWriter,
		*http.Request,
		[]byte,
		*problem_detail.ProblemDetail,
		[]*muxTypes.HeaderEntry,
		error,
	)
	ServerErrorHandler func(
		http.ResponseWriter,
		*http.Request,
		[]byte,
		*problem_detail.ProblemDetail,
		[]*muxTypes.HeaderEntry,
		error,
	)
	SuccessCallback func(*http.Request, []byte, *http.Response, []byte)
	DefaultHeaders  map[string]string
}

func (mux *Mux) ServeHTTP(originalResponseWriter http.ResponseWriter, request *http.Request) {
	if request == nil {
		return
	}

	// Populate the request context.

	if len(mux.SetContextKeyValuePairs) != 0 {
		ctx := request.Context()
		for _, pair := range mux.SetContextKeyValuePairs {
			ctx = context.WithValue(ctx, pair[0], pair[1])
		}
		request = request.WithContext(ctx)
	}

	// Set the client and server error handlers if not specified.

	clientErrorHandler := mux.ClientErrorHandler
	if clientErrorHandler == nil {
		clientErrorHandler = DefaultClientErrorHandler
	}

	serverErrorHandler := mux.ServerErrorHandler
	if serverErrorHandler == nil {
		serverErrorHandler = DefaultServerErrorHandler
	}

	// Use a custom response writer.

	responseWriter := &muxTypes.ResponseWriter{
		ResponseWriter: originalResponseWriter,
		DefaultHeaders: func() map[string]string {
			if defaultHeaders := mux.DefaultHeaders; defaultHeaders != nil {
				return defaultHeaders
			}
			return DefaultHeaders
		}(),
	}

	// Locate the handler.

	requestMethod := strings.ToUpper(request.Method)

	lookupMethod := requestMethod
	if requestMethod == http.MethodHead {
		// A HEAD request is to be processed as if it were a GET request. But signal not write a body.
		lookupMethod = http.MethodGet
		responseWriter.IsHeadRequest = true
	}

	methodToHandlerSpecification, ok := mux.HandlerSpecificationMap[request.URL.Path]
	if !ok {
		clientErrorHandler(
			responseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(http.StatusNotFound, "", nil),
			nil,
			nil,
		)
		return
	}

	handlerSpecification, ok := methodToHandlerSpecification[lookupMethod]
	if !ok {
		expectedMethodsString := strings.Join(slices.Collect(maps.Keys(methodToHandlerSpecification)), ", ")

		clientErrorHandler(
			responseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(
				http.StatusMethodNotAllowed,
				fmt.Sprintf("Expected \"%s\", observed \"%s\"", expectedMethodsString, requestMethod),
				nil,
			),
			[]*muxTypes.HeaderEntry{{Name: "Accept", Value: expectedMethodsString}},
			nil,
		)
		return
	}

	// Perform rate limiting, if specified.

	if rateLimitingConfiguration := handlerSpecification.RateLimitingConfiguration; rateLimitingConfiguration != nil {
		getKeyFunc := rateLimitingConfiguration.GetKey
		if getKeyFunc == nil {
			getKeyFunc = muxTypes.DefaultGetRateLimitingKey
		}

		key, err := getKeyFunc(request)
		if err != nil {
			serverErrorHandler(
				responseWriter,
				request,
				nil,
				nil,
				nil,
				&motmedelErrors.CauseError{
					Message: "An error occurred when extracting a rate limiting key from a request.",
					Cause:   err,
				},
			)
			return
		}

		rateLimitingConfiguration.Lookup.Mutex.Lock()
		if rateLimitingConfiguration.Lookup.Map == nil {
			rateLimitingConfiguration.Lookup.Map = make(map[string]*muxTypes.TimerRateLimiter)
		}
		rateLimitingConfiguration.Lookup.Mutex.Unlock()

		timerRateLimiter, ok := rateLimitingConfiguration.Lookup.Map[key]
		if !ok {
			timerRateLimiter = &muxTypes.TimerRateLimiter{
				RateLimiter: muxTypes.RateLimiter{
					Bucket:           make([]*time.Time, rateLimitingConfiguration.NumRequests),
					NumSecondsExpiry: rateLimitingConfiguration.NumSecondsExpiration,
				},
			}
			rateLimitingConfiguration.Lookup.Map[key] = timerRateLimiter
		}

		if timerRateLimiter.Timer != nil {
			timerRateLimiter.Timer.Stop()
		}
		timerRateLimiter.Timer = time.AfterFunc(
			time.Duration(2*timerRateLimiter.NumSecondsExpiry)*time.Second,
			func() {
				rateLimitingConfiguration.Lookup.Mutex.Lock()
				defer rateLimitingConfiguration.Lookup.Mutex.Unlock()
				if timerRateLimiter.NumOccupied == 0 {
					delete(rateLimitingConfiguration.Lookup.Map, key)
				}
			},
		)

		expirationTime, full := timerRateLimiter.Claim()
		if full {
			clientErrorHandler(
				responseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(http.StatusTooManyRequests, "", nil),
				[]*muxTypes.HeaderEntry{
					{
						Name:  "Retry-After",
						Value: expirationTime.UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT",
					},
				},
				nil,
			)
			return
		}
	}

	// Obtain the body parser configuration.

	bodyParserConfiguration := handlerSpecification.BodyParserConfiguration

	var fullNormalizeContentTypeString string

	// Validate Content-Type

	if bodyParserConfiguration != nil {
		if expectedContentType := bodyParserConfiguration.ContentType; expectedContentType != "" {
			acceptedContentTypeHeaders := []*muxTypes.HeaderEntry{{Name: "Accept", Value: expectedContentType}}

			contentTypeString := request.Header.Get("Content-Type")
			if contentTypeString == "" {
				clientErrorHandler(
					responseWriter,
					request,
					nil,
					problem_detail.MakeStatusCodeProblemDetail(
						http.StatusUnsupportedMediaType,
						"Missing Content-Type",
						nil,
					),
					acceptedContentTypeHeaders,
					nil,
				)
				return
			}

			contentTypeBytes := []byte(contentTypeString)
			contentType, err := content_type.ParseContentType(contentTypeBytes)
			if err != nil {
				serverErrorHandler(
					responseWriter,
					request,
					nil,
					nil,
					nil,
					&motmedelErrors.InputError{
						Message: "An error occurred when attempting to parse the Content-Type header data.",
						Cause:   err,
						Input:   contentTypeBytes,
					},
				)
				return
			}
			if contentType == nil {
				clientErrorHandler(
					responseWriter,
					request,
					nil,
					problem_detail.MakeStatusCodeProblemDetail(
						http.StatusBadRequest,
						"Malformed Content-Type",
						nil,
					),
					nil,
					nil,
				)
				return
			}

			// TODO: The specification could require a certain charset too?
			fullNormalizeContentTypeString = contentType.GetFullType(true)
			if fullNormalizeContentTypeString != expectedContentType {
				clientErrorHandler(
					responseWriter,
					request,
					nil,
					problem_detail.MakeStatusCodeProblemDetail(
						http.StatusUnsupportedMediaType,
						fmt.Sprintf(
							"Expected Content-Type to be \"%s\", observed \"%s\"",
							expectedContentType,
							fullNormalizeContentTypeString,
						),
						nil,
					),
					acceptedContentTypeHeaders,
					nil,
				)
				return
			}
		}
	}

	// Validate Content-Length

	zeroContentLengthStatusCode := http.StatusLengthRequired
	zeroContentLengthMessage := "A body is expected; Content-Length must be set"

	var contentLength uint64
	if _, ok := request.Header["Content-Length"]; ok {
		var err error
		headerValue := request.Header.Get("Content-Length")
		contentLength, err = strconv.ParseUint(headerValue, 10, 64)
		if err != nil {
			clientErrorHandler(
				responseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(
					http.StatusBadRequest,
					"Malformed Content-Length",
					nil,
				),
				nil,
				&motmedelErrors.InputError{
					Message: "An error occurred when attempting to parse the Content-Length as an unsigned integer.",
					Cause:   err,
					Input:   headerValue,
				},
			)
			return
		}
		if contentLength == 0 {
			zeroContentLengthStatusCode = http.StatusBadRequest
			zeroContentLengthMessage = "A body is expected; Content-Length cannot be 0"
		}
	}

	if bodyParserConfiguration != nil && !bodyParserConfiguration.AllowEmpty && contentLength == 0 {
		clientErrorHandler(
			responseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(
				zeroContentLengthStatusCode,
				zeroContentLengthMessage,
				nil,
			),
			nil,
			nil,
		)
		return
	}

	// Obtain the request body

	var requestBody []byte

	if contentLength != 0 {
		var err error
		requestBody, err = io.ReadAll(request.Body)
		if err != nil {
			serverErrorHandler(
				responseWriter,
				request,
				nil,
				nil,
				nil,
				&motmedelErrors.CauseError{
					Message: "An error occurred when reading the request body.",
					Cause:   err,
				},
			)
			return
		}
		defer request.Body.Close()

		if bodyParserConfiguration != nil {
			// NOTE: This should never happen? The Content-Length check should pick this up?
			if !bodyParserConfiguration.AllowEmpty && len(requestBody) == 0 {
				clientErrorHandler(
					responseWriter,
					request,
					nil,
					problem_detail.MakeStatusCodeProblemDetail(
						http.StatusBadRequest,
						"A body is expected",
						nil,
					),
					nil,
					nil,
				)
				return
			}

			switch fullNormalizeContentTypeString {
			case "application/json":
				if !json.Valid(requestBody) {
					clientErrorHandler(
						responseWriter,
						request,
						nil,
						problem_detail.MakeStatusCodeProblemDetail(
							http.StatusBadRequest,
							"Malformed JSON body",
							nil,
						),
						nil,
						nil,
					)
					return
				}
			}

			if parser := bodyParserConfiguration.Parser; parser != nil {
				parsedBody, handlerErrorResponse := parser(request, requestBody)
				if handlerErrorResponse != nil {
					serverError := handlerErrorResponse.ServerError
					clientError := handlerErrorResponse.ClientError
					problemDetail := handlerErrorResponse.ProblemDetail
					headers := handlerErrorResponse.ResponseHeaders

					if serverError != nil {
						serverErrorHandler(responseWriter, request, requestBody, problemDetail, headers, serverError)
						return
					} else if clientError != nil {
						clientErrorHandler(responseWriter, request, requestBody, problemDetail, headers, clientError)
						return
					} else if problemDetail != nil {
						statusCode := problemDetail.Status
						if statusCode >= 400 && statusCode < 500 {
							serverErrorHandler(responseWriter, request, requestBody, problemDetail, headers, nil)
							return
						} else if statusCode >= 500 && statusCode < 600 {
							clientErrorHandler(responseWriter, request, requestBody, problemDetail, headers, nil)
							return
						} else {
							// Assume there were in fact no error?
						}
					}
				}

				request = request.WithContext(
					context.WithValue(request.Context(), ParsedRequestBodyContextKey, parsedBody),
				)
			}
		}
	}

	// Decide whether the response is static content, or requires handling.

	if staticContent := handlerSpecification.StaticContent; staticContent != nil {
		isCached := muxUtils.IfNoneMatchCacheHit(
			request.Header.Get("If-None-Match"),
			staticContent.Etag,
		)
		if !isCached {
			var err error
			isCached, err = muxUtils.IfModifiedSinceCacheHit(
				request.Header.Get("If-Modified-Since"),
				staticContent.LastModified,
			)
			if err != nil {
				if errors.Is(err, muxErrors.ErrBadIfModifiedSinceTimestamp) {
					clientErrorHandler(
						responseWriter,
						request,
						nil,
						problem_detail.MakeBadRequestProblemDetail("Bad If-Modified-Since value", nil),
						nil,
						err,
					)
					return
				} else {
					serverErrorHandler(
						responseWriter,
						request,
						nil,
						nil,
						nil,
						&motmedelErrors.CauseError{
							Message: "An error occurred when checking If-Modified-Since.",
							Cause:   err,
						},
					)
					return
				}
			}
		}

		responseInfo := &muxTypes.ResponseInfo{Headers: staticContent.Headers}
		if isCached {
			responseInfo.StatusCode = http.StatusNotModified
		} else {
			encoding := muxUtils.AcceptContentIdentityIdentifier

			var supportedEncodings []string

			if _, ok := request.Header["Accept-Encoding"]; ok {
				acceptEncoding, err := accept_encoding.ParseAcceptEncoding([]byte(request.Header.Get("Accept-Encoding")))
				if err != nil {
					serverErrorHandler(
						responseWriter,
						request,
						nil,
						nil,
						nil,
						&motmedelErrors.CauseError{
							Message: "An error occurred when parsing the Accept-Encoding header.",
							Cause:   err,
						},
					)
					return
				}
				if acceptEncoding == nil {
					clientErrorHandler(
						responseWriter,
						request,
						nil,
						problem_detail.MakeBadRequestProblemDetail("Bad Accept-Encoding value", nil),
						nil,
						nil,
					)
				}

				supportedEncodings = slices.Collect(maps.Keys(staticContent.ContentEncodingToData))
				contentEncodingToData := staticContent.ContentEncodingToData

				slices.SortFunc(supportedEncodings, func(a, b string) int {
					aData := contentEncodingToData[a].Data
					bData := contentEncodingToData[b].Data
					if len(aData) < len(bData) {
						return -1
					} else if len(aData) > len(bData) {
						return 1
					}
					return 0
				})

				encoding = muxUtils.GetMatchingContentEncoding(
					acceptEncoding.GetPriorityOrderedEncodings(),
					supportedEncodings,
				)
			}

			if encoding == "" {
				clientErrorHandler(
					responseWriter,
					request,
					nil,
					problem_detail.MakeStatusCodeProblemDetail(http.StatusUnsupportedMediaType, "No content encoding could be negotiated", nil),
					[]*muxTypes.HeaderEntry{
						{Name: "Accept-Encoding", Value: strings.Join(supportedEncodings, ", ")},
					},
					nil,
				)
			} else {
				responseInfo.StatusCode = http.StatusOK
				if encoding == muxUtils.AcceptContentIdentityIdentifier {
					responseInfo.Body = staticContent.Data
				} else {
					responseInfo.Headers = append(
						responseInfo.Headers,
						&muxTypes.HeaderEntry{
							Name:  "Content-Encoding",
							Value: encoding,
						},
					)
					responseInfo.Body = staticContent.ContentEncodingToData[encoding].Data
				}
			}
		}

		_ = WriteResponse(responseInfo, responseWriter)
	} else {
		handler := handlerSpecification.Handler
		if handler == nil {
			serverErrorHandler(responseWriter, request, nil, nil, nil, muxErrors.ErrNilHandler)
			return
		}

		responseInfo, handlerErrorResponse := handler(request, requestBody)

		if handlerErrorResponse != nil {
			serverError := handlerErrorResponse.ServerError
			clientError := handlerErrorResponse.ClientError
			problemDetail := handlerErrorResponse.ProblemDetail
			headers := handlerErrorResponse.ResponseHeaders

			if serverError != nil {
				serverErrorHandler(responseWriter, request, requestBody, problemDetail, headers, serverError)
			} else if clientError != nil {
				clientErrorHandler(responseWriter, request, requestBody, problemDetail, headers, clientError)
			} else if problemDetail != nil {
				statusCode := problemDetail.Status
				if statusCode >= 400 && statusCode < 500 {
					serverErrorHandler(responseWriter, request, requestBody, problemDetail, headers, nil)
				} else if statusCode >= 500 && statusCode < 600 {
					clientErrorHandler(responseWriter, request, requestBody, problemDetail, headers, nil)
				} else {
					// The "no response written" error should occur.
				}
			}
		} else if responseInfo != nil {
			_ = WriteResponse(responseInfo, responseWriter)
		} else {
			_ = WriteResponse(&muxTypes.ResponseInfo{}, responseWriter)
		}
	}

	if !responseWriter.WriteHeaderCalled {
		serverErrorHandler(responseWriter, request, requestBody, nil, nil, muxErrors.ErrNoResponseWritten)
	} else {
		if callback := mux.SuccessCallback; callback != nil {
			callback(
				request,
				requestBody,
				&http.Response{StatusCode: responseWriter.WrittenStatusCode, Header: responseWriter.Header()},
				responseWriter.WrittenResponseBody,
			)
		}
	}
}

func (mux *Mux) Add(specifications ...*muxTypes.HandlerSpecification) {
	handlerSpecificationMap := mux.HandlerSpecificationMap
	if handlerSpecificationMap == nil {
		handlerSpecificationMap = make(map[string]map[string]*muxTypes.HandlerSpecification)
	}

	for _, specification := range specifications {
		methodToHandlerSpecification, ok := handlerSpecificationMap[specification.Path]
		if !ok {
			methodToHandlerSpecification = make(map[string]*muxTypes.HandlerSpecification)
			handlerSpecificationMap[specification.Path] = methodToHandlerSpecification
		}

		methodToHandlerSpecification[strings.ToUpper(specification.Method)] = specification
	}

	mux.HandlerSpecificationMap = handlerSpecificationMap
}

func (mux *Mux) Delete(specifications ...*muxTypes.HandlerSpecification) {
	handlerSpecificationMap := mux.HandlerSpecificationMap
	if handlerSpecificationMap == nil {
		return
	}

	for _, specification := range specifications {
		methodToHandlerSpecification, ok := handlerSpecificationMap[specification.Path]
		if !ok {
			return
		}

		delete(methodToHandlerSpecification, strings.ToUpper(specification.Method))

		if len(methodToHandlerSpecification) == 0 {
			delete(handlerSpecificationMap, specification.Path)
		}
	}
}
