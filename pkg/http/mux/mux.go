package mux

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	parseHeaders "github.com/Motmedel/utils_go/pkg/http/parsing/headers"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type WroteHeaderResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (wroteHeaderResponseWriter *WroteHeaderResponseWriter) WriteHeader(statusCode int) {
	wroteHeaderResponseWriter.wroteHeader = true
	wroteHeaderResponseWriter.ResponseWriter.WriteHeader(statusCode)
}

type HandlerSpecification struct {
	Path                string
	Method              string
	ExpectedContentType string
	Handler             func(http.ResponseWriter, *http.Request, []byte) (*problem_detail.ProblemDetail, error, error)
}

func PerformErrorResponse(
	responseWriter http.ResponseWriter,
	request *http.Request,
	problemDetail *problem_detail.ProblemDetail,
	headers [][2]string,
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

	responseWriter.Header().Set("Content-Type", "application/problem+json")
	for _, header := range headers {
		responseWriter.Header().Set(header[0], header[1])
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
	headers [][2]string,
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
	headers [][2]string,
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
	HandlerSpecificationMap map[string]map[string]*HandlerSpecification
	Logger                  *slog.Logger
	DefaultContentType      string
	ClientErrorHandler      func(
		http.ResponseWriter,
		*http.Request,
		[]byte,
		*problem_detail.ProblemDetail,
		[][2]string,
		error,
	)
	ServerErrorHandler func(
		http.ResponseWriter,
		*http.Request,
		[]byte,
		*problem_detail.ProblemDetail,
		[][2]string,
		error,
	)
}

func (mux *Mux) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
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

	wroteHeaderResponseWriter := &WroteHeaderResponseWriter{ResponseWriter: responseWriter}

	methodToHandlerSpecification, ok := mux.HandlerSpecificationMap[request.URL.Path]
	if !ok {
		clientErrorHandler(
			wroteHeaderResponseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(http.StatusNotFound, "", nil),
			nil,
			nil,
		)
		return
	}

	requestMethod := strings.ToUpper(request.Method)
	// "For client requests, an empty string means GET."
	if requestMethod == "" {
		requestMethod = "GET"
	}

	handlerSpecification, ok := methodToHandlerSpecification[requestMethod]
	if !ok {
		expectedMethodsString := strings.Join(slices.Collect(maps.Keys(methodToHandlerSpecification)), ", ")

		clientErrorHandler(
			wroteHeaderResponseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(
				http.StatusMethodNotAllowed,
				fmt.Sprintf("Expected \"%s\", observed \"%s\"", expectedMethodsString, requestMethod),
				nil,
			),
			[][2]string{{"Accept", expectedMethodsString}},
			nil,
		)
		return
	}

	handler := handlerSpecification.Handler
	if handler == nil {
		serverErrorHandler(wroteHeaderResponseWriter, request, nil, nil, nil, muxErrors.ErrNilHandler)
		return
	}

	contentLengthString := request.Header.Get("Content-Length")
	contentLength, err := strconv.Atoi(contentLengthString)
	if err != nil {
		clientErrorHandler(
			wroteHeaderResponseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(http.StatusBadRequest, "Bad Content-Length", nil),
			nil,
			nil,
		)
		return
	}

	if contentLength != 0 {
		expectedContentType := func() string {
			if handlerSpecification.ExpectedContentType != "" {
				return handlerSpecification.ExpectedContentType
			}
			return mux.DefaultContentType
		}()
		headers := [][2]string{{"Accept", expectedContentType}}

		contentTypeString := request.Header.Get("Content-Type")
		if contentTypeString == "" {
			clientErrorHandler(
				wroteHeaderResponseWriter,
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

		contentTypeBytes := []byte(contentTypeString)
		contentType, err := parseHeaders.ParseContentType(contentTypeBytes)
		if err != nil {
			clientErrorHandler(
				wroteHeaderResponseWriter,
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
				wroteHeaderResponseWriter,
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

		body, err := io.ReadAll(request.Body)
		if err != nil {
			serverErrorHandler(wroteHeaderResponseWriter, request, nil, nil, nil, err)
			return
		}
		defer request.Body.Close()

		// TODO: Handler cannot currently control the header values in the error response, but should be able to?

		problemDetail, clientError, serverError := handler(wroteHeaderResponseWriter, request, body)
		if serverError != nil {
			serverErrorHandler(wroteHeaderResponseWriter, request, body, problemDetail, nil, serverError)
		} else if clientError != nil {
			clientErrorHandler(wroteHeaderResponseWriter, request, body, problemDetail, nil, clientError)
		} else {
			// TODO: If problemDetail is not nil but no client or server error, the problem detail should be inspected
			//	and the corresponding error handler function should be called.
			if !wroteHeaderResponseWriter.wroteHeader {
				serverErrorHandler(wroteHeaderResponseWriter, request, body, nil, nil, muxErrors.ErrNoResponseWritten)
			}
		}
	}
}

func (mux *Mux) Add(specifications ...*HandlerSpecification) {
	handlerSpecificationMap := mux.HandlerSpecificationMap
	if handlerSpecificationMap == nil {
		handlerSpecificationMap = make(map[string]map[string]*HandlerSpecification)
	}

	for _, specification := range specifications {
		methodToHandlerSpecification, ok := handlerSpecificationMap[specification.Path]
		if !ok {
			methodToHandlerSpecification = make(map[string]*HandlerSpecification)
			handlerSpecificationMap[specification.Path] = methodToHandlerSpecification
		}

		methodToHandlerSpecification[strings.ToUpper(specification.Method)] = specification
	}

	mux.HandlerSpecificationMap = handlerSpecificationMap
}

func (mux *Mux) Delete(specifications ...*HandlerSpecification) {
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
