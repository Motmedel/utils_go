package mux

import (
	"context"
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
)

var DefaultHeaders = map[string]string{
	"Cache-Control":                "no-store",
	"Cross-Origin-Opener-Policy":   "same-origin",
	"Cross-Origin-Embedder-Policy": "require-cors",
	"Cross-Origin-Resource-Policy": "same-origin",
	"Content-Security-Policy":      "default-src 'self'; frame-ancestors 'none'; base-uri 'none', form-action 'none'",
	"X-Content-Type-Options":       "nosniff",
	"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
}

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
	responseWriterHeader.Set("Content-Type", "application/problem+json")
	for _, header := range responseInfo.Headers {
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
	DefaultHeaders map[string]string
}

func (mux *Mux) ServeHTTP(originalResponseWriter http.ResponseWriter, request *http.Request) {
	if request == nil {
		return
	}

	if len(mux.SetContextKeyValuePairs) != 0 {
		ctx := request.Context()
		for _, pair := range mux.SetContextKeyValuePairs {
			ctx = context.WithValue(ctx, pair[0], pair[1])
		}
		request = request.WithContext(ctx)
	}

	clientErrorHandler := mux.ClientErrorHandler
	if clientErrorHandler == nil {
		clientErrorHandler = DefaultClientErrorHandler
	}

	serverErrorHandler := mux.ServerErrorHandler
	if serverErrorHandler == nil {
		serverErrorHandler = DefaultServerErrorHandler
	}

	responseWriter := &muxTypes.ResponseWriter{
		ResponseWriter: originalResponseWriter,
		DefaultHeaders: func() map[string]string {
			if defaultHeaders := mux.DefaultHeaders; defaultHeaders != nil {
				return defaultHeaders
			}
			return DefaultHeaders
		}(),
	}

	requestMethod := strings.ToUpper(request.Method)
	// "For client requests, an empty string means GET."
	if requestMethod == "" {
		requestMethod = "GET"
	}

	lookupMethod := requestMethod
	if requestMethod == "HEAD" {
		lookupMethod = "GET"
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
			[]*muxTypes.HeaderEntry{
				{Name: "Accept", Value: expectedMethodsString},
			},
			nil,
		)
		return
	}

	var contentLength int
	if _, ok := request.Header["Content-Length"]; ok {
		var err error
		contentLength, err = strconv.Atoi(request.Header.Get("Content-Length"))
		if err != nil {
			clientErrorHandler(
				responseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(http.StatusBadRequest, "Bad Content-Length", nil),
				nil,
				nil,
			)
			return
		}
	}

	var requestBody []byte

	if contentLength != 0 {
		expectedContentType := func() string {
			if handlerSpecification.ExpectedContentType != "" {
				return handlerSpecification.ExpectedContentType
			}
			return ""
		}()
		headers := []*muxTypes.HeaderEntry{{Name: "Accept", Value: expectedContentType}}

		contentTypeString := request.Header.Get("Content-Type")
		if contentTypeString == "" {
			clientErrorHandler(
				responseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnsupportedMediaType,
					"Content-Type is not set",
					nil,
				),
				headers,
				nil,
			)
			return
		}

		contentType, err := content_type.ParseContentType([]byte(contentTypeString))
		if err != nil {
			clientErrorHandler(
				responseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(http.StatusBadRequest, "Bad Content-Type", nil),
				nil,
				&motmedelErrors.CauseError{
					Message: "An error occurred when attempting to parse the Content-Type header data.",
					Cause:   err,
				},
			)
			return
		}

		// TODO: The specification could require a certain charset too?
		fullNormalizeContentTypeString := contentType.GetFullType(true)
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
				headers,
				nil,
			)
			return
		}

		requestBody, err = io.ReadAll(request.Body)
		if err != nil {
			serverErrorHandler(responseWriter, request, nil, nil, nil, err)
			return
		}
		defer request.Body.Close()
	}

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
					serverErrorHandler(responseWriter, request, nil, nil, nil, err)
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
					serverErrorHandler(responseWriter, request, nil, nil, nil, err)
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

	if !responseWriter.WriteHeaderCaller {
		serverErrorHandler(responseWriter, request, requestBody, nil, nil, muxErrors.ErrNoResponseWritten)
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
