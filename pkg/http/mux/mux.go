package mux

import (
	"fmt"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	parseHeaders "github.com/Motmedel/utils_go/pkg/http/parsing/headers"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	"io"
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
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	if problemDetail.Status == 0 {
		motmedelLog.LogError(
			"The error response problem detail status is unset.",
			muxErrors.ErrUnsetErrorResponseProblemDetailStatus,
			logger,
		)
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
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

func DefaultBadRequestHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	requestBody []byte,
	problemDetail *problem_detail.ProblemDetail,
	headers [][2]string,
) {
	if problemDetail != nil {
		problemDetail = problem_detail.MakeBadRequestProblemDetail("", nil)
	}
	PerformErrorResponse(responseWriter, request, problemDetail, headers)
}

func DefaultInternalServerErrorHandler(
	responseWriter http.ResponseWriter,
	request *http.Request,
	requestBody []byte,
	err error,
) {
	motmedelLog.LogError(
		"A server error occurred.",
		err,
		motmedelLog.GetLoggerFromCtxWithDefault(request.Context(), nil),
	)

	PerformErrorResponse(
		responseWriter,
		request,
		problem_detail.MakeInternalServerErrorProblemDetail("", nil),
		nil,
	)
}

type Mux struct {
	HandlerSpecificationMap map[string]map[string]*HandlerSpecification
	DefaultContentType      string
	BadRequestHandler       func(
		http.ResponseWriter,
		*http.Request,
		[]byte,
		*problem_detail.ProblemDetail,
		[][2]string,
	)
	InternalServerErrorHandler func(
		http.ResponseWriter,
		*http.Request,
		[]byte,
		error,
	)
}

func (mux *Mux) ServeHttp(responseWriter http.ResponseWriter, request *http.Request) {
	badRequestHandler := mux.BadRequestHandler
	if badRequestHandler == nil {
		badRequestHandler = DefaultBadRequestHandler
	}

	internalServerErrorHandler := mux.InternalServerErrorHandler
	if internalServerErrorHandler == nil {
		internalServerErrorHandler = DefaultInternalServerErrorHandler
	}

	wroteHeaderResponseWriter := &WroteHeaderResponseWriter{ResponseWriter: responseWriter}

	methodToHandlerSpecification, ok := mux.HandlerSpecificationMap[request.URL.Path]
	if !ok {
		badRequestHandler(
			wroteHeaderResponseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(http.StatusNotFound, "", nil),
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

		badRequestHandler(
			wroteHeaderResponseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(
				http.StatusMethodNotAllowed,
				fmt.Sprintf("Expected \"%s\", observed \"%s\"", expectedMethodsString, requestMethod),
				nil,
			),
			[][2]string{{"Accept", expectedMethodsString}},
		)
		return
	}

	handler := handlerSpecification.Handler
	if handler == nil {
		internalServerErrorHandler(wroteHeaderResponseWriter, request, nil, muxErrors.ErrNilHandler)
		return
	}

	contentLengthString := request.Header.Get("Content-Length")
	contentLength, err := strconv.Atoi(contentLengthString)
	if err != nil {
		badRequestHandler(
			wroteHeaderResponseWriter,
			request,
			nil,
			problem_detail.MakeStatusCodeProblemDetail(http.StatusBadRequest, "Bad Content-Length", nil),
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
			badRequestHandler(
				wroteHeaderResponseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnsupportedMediaType,
					"Content-Type is not set",
					nil,
				),
				headers,
			)
			return
		}

		contentType, err := parseHeaders.ParseContentType([]byte(contentTypeString))
		if err != nil {
			// TODO: Do something with the error.
			badRequestHandler(
				wroteHeaderResponseWriter,
				request,
				nil,
				problem_detail.MakeStatusCodeProblemDetail(http.StatusBadRequest, "Bad Content-Type", nil),
				nil,
			)
			return
		}

		fullNormalizeContentTypeString := contentType.GetFullType(true)
		if fullNormalizeContentTypeString != expectedContentType {
			badRequestHandler(
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
			)
			return
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			internalServerErrorHandler(wroteHeaderResponseWriter, request, nil, err)
			return
		}
		defer request.Body.Close()

		problemDetail, clientError, serverError := handler(wroteHeaderResponseWriter, request, body)
		if serverError != nil {
			internalServerErrorHandler(wroteHeaderResponseWriter, request, body, serverError)
		} else if clientError != nil {
			badRequestHandler(wroteHeaderResponseWriter, request, body, problemDetail, nil)
		} else {
			if !wroteHeaderResponseWriter.wroteHeader {
				internalServerErrorHandler(wroteHeaderResponseWriter, request, body, muxErrors.ErrNoResponseWritten)
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
			methodToHandlerSpecification = handlerSpecificationMap[specification.Path]
			handlerSpecificationMap[specification.Path] = methodToHandlerSpecification
		}

		methodToHandlerSpecification[strings.ToUpper(specification.Method)] = specification
	}
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

		if len(methodToHandlerSpecification) == 9 {
			delete(handlerSpecificationMap, specification.Path)
		}
	}
}
