package mux

import (
	"context"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypes "github.com/Motmedel/utils_go/pkg/http/mux/types"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var DefaultHeaders = map[string]string{
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
	var skippedDefaultHeadersSet map[string]struct{}

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

	responseWriter.WriteHeader(responseInfo.StatusCode)
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
	headers []*muxTypes.HeaderEntry,
) {
	logger := motmedelLog.GetLoggerFromCtxWithDefault(request.Context(), nil)

	if problemDetail == nil {
		motmedelLog.LogError(
			"The error response problem detail is nil.",
			muxErrors.ErrNilErrorResponseProblemDetail,
			logger,
		)
		problemDetail = problem_detail.MakeInternalServerErrorProblemDetail("", nil)
		headers = nil
	}

	if problemDetail.Status == 0 {
		motmedelLog.LogError(
			"The error response problem detail status is unset.",
			muxErrors.ErrUnsetErrorResponseProblemDetailStatus,
			logger,
		)
		problemDetail = problem_detail.MakeInternalServerErrorProblemDetail("", nil)
		headers = nil
	}

	problemDetailString, err := problemDetail.String()
	if err != nil {
		motmedelLog.LogError(
			"An error occurred when converting a problem detail into a string.",
			err,
			logger,
		)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	var defaultHeaders map[string]string
	muxResponseWriter, ok := responseWriter.(*muxTypes.ResponseWriter)
	if ok {
		defaultHeaders = muxResponseWriter.DefaultHeaders
	} else {
		defaultHeaders = make(map[string]string)
	}
	var skippedDefaultHeadersSet map[string]struct{}

	responseWriterHeader := responseWriter.Header()
	responseWriterHeader.Set("Content-Type", "application/problem+json")
	for _, header := range headers {
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

	responseWriter.WriteHeader(problemDetail.Status)

	if _, err = io.WriteString(responseWriter, problemDetailString); err != nil {
		motmedelLog.LogError(
			"An error occurred when writing the HTTP response body.",
			err,
			logger,
		)
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
	headers []*muxTypes.HeaderEntry,
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
	PerformErrorResponse(responseWriter, request, problemDetail, headers)
}

type Mux struct {
	HandlerSpecificationMap map[string]map[string]*muxTypes.HandlerSpecification
	Logger                  *slog.Logger
	DefaultContentType      string
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

	request = request.WithContext(motmedelLog.CtxWithLogger(request.Context(), mux.Logger))

	clientErrorHandler := mux.ClientErrorHandler
	if clientErrorHandler == nil {
		clientErrorHandler = DefaultClientErrorHandler
	}

	serverErrorHandler := mux.ServerErrorHandler
	if serverErrorHandler == nil {
		serverErrorHandler = DefaultServerErrorHandler
	}

	customResponseWriter := &muxTypes.ResponseWriter{ResponseWriter: originalResponseWriter}

	defaultHeaders := mux.DefaultHeaders
	if defaultHeaders == nil {
		defaultHeaders = DefaultHeaders
	}
	customResponseWriter.DefaultHeaders = defaultHeaders

	requestMethod := strings.ToUpper(request.Method)
	// "For client requests, an empty string means GET."
	if requestMethod == "" {
		requestMethod = "GET"
	}

	lookupMethod := requestMethod
	if requestMethod == "HEAD" {
		lookupMethod = "GET"
		customResponseWriter.IsHeadRequest = true
	}

	methodToHandlerSpecification, ok := mux.HandlerSpecificationMap[request.URL.Path]
	if !ok {
		clientErrorHandler(
			customResponseWriter,
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
			customResponseWriter,
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
				customResponseWriter,
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
			return mux.DefaultContentType
		}()
		headers := []*muxTypes.HeaderEntry{{Name: "Accept", Value: expectedContentType}}

		contentTypeString := request.Header.Get("Content-Type")
		if contentTypeString == "" {
			clientErrorHandler(
				customResponseWriter,
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
				customResponseWriter,
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
				customResponseWriter,
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
			serverErrorHandler(customResponseWriter, request, nil, nil, nil, err)
			return
		}
		defer request.Body.Close()
	}

	if len(mux.SetContextKeyValuePairs) != 0 {
		ctx := request.Context()
		for _, pair := range mux.SetContextKeyValuePairs {
			ctx = context.WithValue(ctx, pair[0], pair[1])
		}
		request = request.WithContext(ctx)
	}

	if staticContent := handlerSpecification.StaticContent; staticContent != nil {
		isCached := func() bool {
			if ifNoneMatch := request.Header.Get("If-None-Match"); ifNoneMatch != "" && staticContent.Etag != "" {
				return ifNoneMatch == staticContent.Etag
			}

			if ifModifiedSince := request.Header.Get("If-Modified-Since"); ifModifiedSince != "" && staticContent.LastModified != "" {
				// TODO: Perform timestamp check.
				return false
			}

			return false
		}()

		for _, header := range staticContent.Headers {
			customResponseWriter.Header().Set(header.Name, header.Value)
		}

		if isCached {
			customResponseWriter.WriteHeader(http.StatusNotModified)
		} else {
			if _, err := customResponseWriter.Write(staticContent.Data); err != nil {
				serverErrorHandler(customResponseWriter, request, nil, nil, nil, err)
			}
		}
	} else {
		handler := handlerSpecification.Handler
		if handler == nil {
			serverErrorHandler(customResponseWriter, request, nil, nil, nil, muxErrors.ErrNilHandler)
			return
		}

		muxResponse, handlerErrorResponse := handler(request, requestBody)

		if handlerErrorResponse != nil {
			serverError := handlerErrorResponse.ServerError
			clientError := handlerErrorResponse.ClientError
			problemDetail := handlerErrorResponse.ProblemDetail
			headers := handlerErrorResponse.ResponseHeaders

			if serverError != nil {
				serverErrorHandler(customResponseWriter, request, requestBody, problemDetail, headers, serverError)
			} else if clientError != nil {
				clientErrorHandler(customResponseWriter, request, requestBody, problemDetail, headers, clientError)
			} else if problemDetail != nil {
				statusCode := problemDetail.Status
				if statusCode >= 400 && statusCode < 500 {
					serverErrorHandler(customResponseWriter, request, requestBody, problemDetail, headers, nil)
				} else if statusCode >= 500 && statusCode < 600 {
					clientErrorHandler(customResponseWriter, request, requestBody, problemDetail, headers, nil)
				} else {
					// The "no response written" error should occur.
				}
			}
		} else if muxResponse != nil {

		} else {
			// The "no response written" error should occur.
		}
	}

	if !customResponseWriter.WriteHeaderCaller {
		serverErrorHandler(customResponseWriter, request, requestBody, nil, nil, muxErrors.ErrNoResponseWritten)
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
