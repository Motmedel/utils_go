package mux

import (
	"context"
	"encoding/json"
	"fmt"
	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxContext "github.com/Motmedel/utils_go/pkg/http/mux/context"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxInternal "github.com/Motmedel/utils_go/pkg/http/mux/internal"
	muxInternalMux "github.com/Motmedel/utils_go/pkg/http/mux/internal/mux"
	muxTypesEnpointSpecification "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	muxTypesFirewall "github.com/Motmedel/utils_go/pkg/http/mux/types/firewall"
	muxTypesMiddleware "github.com/Motmedel/utils_go/pkg/http/mux/types/middleware"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponseWriter "github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser"
	muxUtilsContentNegotiation "github.com/Motmedel/utils_go/pkg/http/mux/utils/content_negotiation"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"strings"
)

type baseMux struct {
	SetContextKeyValuePairs [][2]any
	ResponseErrorHandler    func(context.Context, *muxTypesResponseError.ResponseError, *muxTypesResponseWriter.ResponseWriter)
	DoneCallback            func(context.Context)
	FirewallConfiguration   *muxTypesFirewall.Configuration
	DefaultHeaders          map[string]string
	DefaultDocumentHeaders  map[string]string
	Middleware              []muxTypesMiddleware.Middleware
	ProblemDetailConverter  muxTypesResponseError.ProblemDetailConverter
}

func (bm *baseMux) getFirewallVerdict(request *http.Request) (muxTypesFirewall.Verdict, *muxTypesResponseError.ResponseError) {
	if request == nil {
		return muxTypesFirewall.VerdictReject, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	if firewallConfiguration := bm.FirewallConfiguration; firewallConfiguration != nil {
		if firewallHandler := firewallConfiguration.Handler; firewallHandler != nil {
			return firewallHandler(request)
		}
	}

	return muxTypesFirewall.VerdictAccept, nil
}

func (bm *baseMux) ServeHttpWithCallback(
	originalResponseWriter http.ResponseWriter,
	request *http.Request,
	callback func(*http.Request, *muxTypesResponseWriter.ResponseWriter) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError),
) {
	if originalResponseWriter == nil {
		return
	}

	if request == nil {
		return
	}

	if callback == nil {
		return
	}

	responseErrorHandler := bm.ResponseErrorHandler
	if responseErrorHandler == nil {
		responseErrorHandler = muxInternal.DefaultResponseErrorHandler
	}

	// Populate the request context.

	httpContext := &motmedelHttpTypes.HttpContext{Request: request}
	request = request.WithContext(
		context.WithValue(request.Context(), motmedelHttpContext.HttpContextContextKey, httpContext),
	)

	requestId, err := uuid.NewV7()
	if err != nil {
		slog.WarnContext(
			motmedelContext.WithErrorContextValue(
				request.Context(),
				motmedelErrors.NewWithTrace(fmt.Errorf("uuid new v7: %w", err)),
			),
			"An error occurred when generating a request id.",
		)
	} else {
		contextRequest := request.WithContext(
			context.WithValue(request.Context(), motmedelHttpContext.RequestIdContextKey, requestId.String()),
		)
		if contextRequest != nil {
			request = contextRequest
		}
	}

	if len(bm.SetContextKeyValuePairs) != 0 {
		ctx := request.Context()
		for _, pair := range bm.SetContextKeyValuePairs {
			ctx = context.WithValue(ctx, pair[0], pair[1])
		}
		if contextRequest := request.WithContext(ctx); contextRequest != nil {
			request = contextRequest
		}
	}

	// Use a custom response writer.

	var responseWriter *muxTypesResponseWriter.ResponseWriter

	if convertedResponseWriter, ok := originalResponseWriter.(*muxTypesResponseWriter.ResponseWriter); ok {
		responseWriter = convertedResponseWriter
		originalResponseWriter = convertedResponseWriter.ResponseWriter

		responseWriter.DefaultHeaders = bm.DefaultHeaders
		responseWriter.DefaultDocumentHeaders = bm.DefaultDocumentHeaders
	} else {
		responseWriter = &muxTypesResponseWriter.ResponseWriter{
			ResponseWriter:         originalResponseWriter,
			DefaultHeaders:         bm.DefaultHeaders,
			DefaultDocumentHeaders: bm.DefaultDocumentHeaders,
		}
	}

	responseWriter.IsHeadRequest = strings.ToUpper(request.Method) == http.MethodHead

	// Check the request with the muxTypesFirewall.

	verdict, firewallResponseError := bm.getFirewallVerdict(request)
	if verdict == muxTypesFirewall.VerdictDrop {
		hijacker, ok := originalResponseWriter.(http.Hijacker)
		if ok {
			connection, _, err := hijacker.Hijack()
			if err != nil {
				responseErrorHandler(
					request.Context(),
					&muxTypesResponseError.ResponseError{
						ServerError: motmedelErrors.NewWithTrace(
							fmt.Errorf("response writer hijacker hijack: %w", err),
						),
					},
					responseWriter,
				)
			}
			if connection != nil {
				if err := connection.Close(); err != nil {
					slog.ErrorContext(
						motmedelContext.WithErrorContextValue(
							request.Context(),
							motmedelErrors.NewWithTrace(
								fmt.Errorf("connection close: %w", err),
							),
						),
						"An error occurred when closing a connection.",
					)
				}
			}
			return
		} else {
			// Trigger a termination of the connection.
			panic(http.ErrAbortHandler)
		}
	} else if verdict == muxTypesFirewall.VerdictReject {
		if firewallResponseError == nil {
			firewallResponseError = &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(http.StatusForbidden, "", nil),
			}
		}
		responseErrorHandler(request.Context(), firewallResponseError, responseWriter)
	} else {
		for _, middleware := range bm.Middleware {
			if middleware != nil {
				if middlewareRequest := middleware(request); middlewareRequest != nil {
					request = middlewareRequest
				}
			}
		}

		var acceptEncoding *motmedelHttpTypes.AcceptEncoding

		if contentNegotiation, _ := muxUtilsContentNegotiation.GetContentNegotiation(request.Header, false); contentNegotiation != nil {
			request = request.WithContext(
				context.WithValue(request.Context(), muxContext.ContentNegotiationContextKey, contentNegotiation),
			)
			acceptEncoding = contentNegotiation.AcceptEncoding
		}

		// Respond to the request.

		response, responseError := callback(request, responseWriter)

		if !responseWriter.WriteHeaderCalled {
			if responseError != nil {
				responseErrorHandler(request.Context(), responseError, responseWriter)
			} else {
				if response == nil {
					response = &muxTypesResponse.Response{}
				}

				if err := responseWriter.WriteResponse(request.Context(), response, acceptEncoding); err != nil {
					responseErrorHandler(
						request.Context(),
						&muxTypesResponseError.ResponseError{
							ServerError: motmedelErrors.MakeError(
								fmt.Errorf("write response: %w", err),
								response,
							),
						},
						responseWriter,
					)
				}
			}
		}

		httpContext.Response = &http.Response{
			StatusCode: responseWriter.WrittenStatusCode,
			Header:     responseWriter.Header(),
		}
		httpContext.ResponseBody = responseWriter.WrittenBody
	}

	// Handle the case when no response was produced.

	if !responseWriter.WriteHeaderCalled {
		responseErrorHandler(
			request.Context(),
			&muxTypesResponseError.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNoResponseWritten),
			},
			responseWriter,
		)
	}

	if doneCallback := bm.DoneCallback; doneCallback != nil {
		doneCallback(request.Context())
	}
}

type Mux struct {
	baseMux
	HandlerSpecificationMap map[string]map[string]*muxTypesEnpointSpecification.EndpointSpecification
}

func muxHandleRequest(
	mux *Mux,
	request *http.Request,
	responseWriter http.ResponseWriter,
) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
	if mux == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilMux),
		}
	}

	if request == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestHeader := request.Header
	if requestHeader == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	httpContext, ok := request.Context().Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext)
	if !ok {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrCouldNotObtainHttpContext),
		}
	}
	if httpContext == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpContext),
		}
	}

	// Locate the endpoint specification.

	endpointSpecification, responseInfo, responseError := muxInternalMux.ObtainEndpointSpecification(
		mux.HandlerSpecificationMap,
		request,
	)
	if responseInfo != nil || responseError != nil {
		return responseInfo, responseError
	}
	if endpointSpecification == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilEndpointSpecification),
		}
	}

	// Perform rate limiting, if specified.

	if rateLimitingConfiguration := endpointSpecification.RateLimitingConfiguration; rateLimitingConfiguration != nil {
		if responseError := muxInternalMux.HandleRateLimiting(rateLimitingConfiguration, request); responseError != nil {
			return nil, responseError
		}
	}

	// TODO: Obtain the parsed url.

	// Obtain the parsed header.

	if headerParserConfiguration := endpointSpecification.HeaderParserConfiguration; headerParserConfiguration != nil {
		if parser := headerParserConfiguration.Parser; parser != nil {
			parsedHeader, responseError := parser(request)
			if responseError != nil {
				return nil, responseError
			}

			request = request.WithContext(
				context.WithValue(request.Context(), parsing.ParsedRequestHeaderContextKey, parsedHeader),
			)
		}
	}

	// Validate body parameters, and obtain and validate the body

	allowEmptyBody := true
	var expectedContentType string
	var bodyParser body_parser.BodyParser
	var maxBytes int64

	// Obtain validation options from the handler specification configuration.
	if bodyParserConfiguration := endpointSpecification.BodyParserConfiguration; bodyParserConfiguration != nil {
		allowEmptyBody = bodyParserConfiguration.AllowEmpty
		expectedContentType = bodyParserConfiguration.ContentType
		bodyParser = bodyParserConfiguration.Parser
		maxBytes = bodyParserConfiguration.MaxBytes
	}

	// Validate Content-Type (parse and match header value against accepted value)
	if expectedContentType != "" {
		if responseError = muxInternalMux.ValidateContentType(expectedContentType, requestHeader); responseError != nil {
			return nil, responseError
		}
	}

	if maxBytes > 0 {
		request.Body = http.MaxBytesReader(responseWriter, request.Body, maxBytes)
	}

	// Validate Content-Length (parse and check if empty is accepted)
	if responseError := muxInternalMux.ValidateContentLength(allowEmptyBody, requestHeader); responseError != nil {
		return nil, responseError
	}

	// Obtain the request body
	requestBody, responseError := muxInternalMux.ObtainRequestBody(request.ContentLength, request.Body, maxBytes)
	if responseError != nil {
		return nil, responseError
	}
	httpContext.RequestBody = requestBody

	if !allowEmptyBody && len(requestBody) == 0 {
		return nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusBadRequest,
				"A body is expected.",
				nil,
			),
		}
	}

	// Basic check to see if the request body conforms to the expected content type.
	switch expectedContentType {
	case "application/json":
		if !json.Valid(requestBody) {
			return nil, &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusBadRequest,
					"Invalid JSON body.",
					nil,
				),
			}
		}
	}

	// Parse the body.
	if bodyParser != nil {
		parsedBody, responseError := bodyParser.Parse(request, requestBody)
		if responseError != nil {
			return nil, responseError
		}

		request = request.WithContext(
			context.WithValue(request.Context(), parsing.ParsedRequestBodyContextKey, parsedBody),
		)
	}

	// Respond with static content.

	if staticContent := endpointSpecification.StaticContent; staticContent != nil {
		isCached, responseError := muxInternalMux.ObtainIsCached(staticContent, requestHeader)
		if responseError != nil {
			return nil, responseError
		}

		var acceptEncoding *motmedelHttpTypes.AcceptEncoding
		contentNegotiation, _ := request.Context().Value(muxContext.ContentNegotiationContextKey).(*motmedelHttpTypes.ContentNegotiation)
		if contentNegotiation != nil {
			acceptEncoding = contentNegotiation.AcceptEncoding
		}

		return muxInternalMux.ObtainStaticContentResponse(staticContent, isCached, requestHeader, acceptEncoding)
	}

	// Respond with dynamic content (via a handler).

	if handler := endpointSpecification.Handler; handler != nil {
		return handler(request, requestBody)
	}

	return nil, nil

}

func (mux *Mux) ServeHTTP(originalResponseWriter http.ResponseWriter, request *http.Request) {
	mux.baseMux.ServeHttpWithCallback(
		originalResponseWriter,
		request,
		func(request *http.Request, responseWriter *muxTypesResponseWriter.ResponseWriter) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
			response, responseError := muxHandleRequest(mux, request, responseWriter)
			if responseError != nil {
				responseError.ProblemDetailConverter = mux.ProblemDetailConverter
			}

			if responseWriter != nil {
				responseWriter.DefaultHeaders = mux.DefaultHeaders
				responseWriter.DefaultDocumentHeaders = mux.DefaultDocumentHeaders
			}

			return response, responseError
		},
	)
}

func (mux *Mux) Add(specifications ...*muxTypesEnpointSpecification.EndpointSpecification) {
	if len(specifications) == 0 {
		return
	}

	handlerSpecificationMap := mux.HandlerSpecificationMap
	if handlerSpecificationMap == nil {
		handlerSpecificationMap = make(map[string]map[string]*muxTypesEnpointSpecification.EndpointSpecification)
	}

	for _, specification := range specifications {
		methodToHandlerSpecification, ok := handlerSpecificationMap[specification.Path]
		if !ok {
			methodToHandlerSpecification = make(map[string]*muxTypesEnpointSpecification.EndpointSpecification)
			handlerSpecificationMap[specification.Path] = methodToHandlerSpecification
		}

		methodToHandlerSpecification[strings.ToUpper(specification.Method)] = specification
	}

	mux.HandlerSpecificationMap = handlerSpecificationMap
}

func (mux *Mux) Delete(specifications ...*muxTypesEnpointSpecification.EndpointSpecification) {
	if len(specifications) == 0 {
		return
	}

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
