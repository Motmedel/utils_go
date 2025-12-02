package mux

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxContext "github.com/Motmedel/utils_go/pkg/http/mux/context"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser"
	muxInternal "github.com/Motmedel/utils_go/pkg/http/mux/internal"
	muxInternalMux "github.com/Motmedel/utils_go/pkg/http/mux/internal/mux"
	muxTypesEnpointSpecification "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	muxTypesFirewall "github.com/Motmedel/utils_go/pkg/http/mux/types/firewall"
	muxTypesMiddleware "github.com/Motmedel/utils_go/pkg/http/mux/types/middleware"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponseWriter "github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	muxUtilsContentNegotiation "github.com/Motmedel/utils_go/pkg/http/mux/utils/content_negotiation"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelIter "github.com/Motmedel/utils_go/pkg/iter"
	"github.com/Motmedel/utils_go/pkg/utils"
	"github.com/google/uuid"
)

type muxHttpContextContextType struct{}

var MuxHttpContextContextKey muxHttpContextContextType

// TODO: Do all of these need to be here, or can they be moved the the `Mux` struct?t
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

	// Create an HTTP context and populate it with the request and put it in the request context.

	httpContext := &motmedelHttpTypes.HttpContext{Request: request}
	request = request.WithContext(
		context.WithValue(request.Context(), MuxHttpContextContextKey, httpContext),
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
					motmedelHttpContext.WithHttpContextValue(request.Context(), httpContext),
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
		responseErrorHandler(motmedelHttpContext.WithHttpContextValue(request.Context(), httpContext), firewallResponseError, responseWriter)
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
				responseErrorHandler(motmedelHttpContext.WithHttpContextValue(request.Context(), httpContext), responseError, responseWriter)
			} else {
				if response == nil {
					response = &muxTypesResponse.Response{}
				}

				if err := responseWriter.WriteResponse(request.Context(), response, acceptEncoding); err != nil {
					responseErrorHandler(
						motmedelHttpContext.WithHttpContextValue(request.Context(), httpContext),
						&muxTypesResponseError.ResponseError{
							ServerError: motmedelErrors.New(
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
			motmedelHttpContext.WithHttpContextValue(request.Context(), httpContext),
			&muxTypesResponseError.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNoResponseWritten),
			},
			responseWriter,
		)
	}

	if doneCallback := bm.DoneCallback; doneCallback != nil {
		doneCallback(motmedelHttpContext.WithHttpContextValue(request.Context(), httpContext))
	}
}

type Mux struct {
	baseMux
	EndpointSpecificationMap map[string]map[string]*muxTypesEnpointSpecification.EndpointSpecification
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

	httpContext, ok := request.Context().Value(MuxHttpContextContextKey).(*motmedelHttpTypes.HttpContext)
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

	endpointSpecification, methodToEndpointSpecification, responseError := muxInternalMux.GetEndpointSpecification(
		mux.EndpointSpecificationMap,
		request,
	)
	if responseError != nil {
		return nil, responseError
	}

	// There exists no endpoint specification for the given method,
	if endpointSpecification == nil {
		// and for no other methods either, which is an error (as 404 should be produced by `GetEndpointSpecification`)
		if len(methodToEndpointSpecification) == 0 {
			return nil, &muxTypesResponseError.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilEndpointSpecification),
			}
		}

		// Produce an OPTIONS response (list allowed methods and/or CORS configuration).

		var allowedMethods []string
		var corsEndpointSpecifications []*muxTypesEnpointSpecification.EndpointSpecification

		for method, otherEndpointSpecification := range methodToEndpointSpecification {
			if otherEndpointSpecification == nil {
				continue
			}

			allowedMethods = append(allowedMethods, method)

			if corsRequestParser := otherEndpointSpecification.CorsRequestParser; !utils.IsNil(corsRequestParser) {
				corsEndpointSpecifications = append(corsEndpointSpecifications, otherEndpointSpecification)
			}
		}

		if _, ok := methodToEndpointSpecification[http.MethodHead]; !ok {
			if _, ok := methodToEndpointSpecification[http.MethodGet]; ok {
				allowedMethods = append(allowedMethods, http.MethodHead)
			}
		}
		if _, ok := methodToEndpointSpecification[http.MethodOptions]; !ok {
			allowedMethods = append(allowedMethods, http.MethodOptions)
		}
		slices.Sort(allowedMethods)

		expectedMethodsString := strings.Join(allowedMethods, ", ")
		headerEntries := []*muxTypesResponse.HeaderEntry{{Name: "Allow", Value: expectedMethodsString}}

		if strings.ToUpper(request.Method) == http.MethodOptions {
			if len(corsEndpointSpecifications) > 0 {
				var corsConfiguration motmedelHttpTypes.CorsConfiguration
				accessControlRequestMethod := strings.ToUpper(requestHeader.Get("Access-Control-Request-Method"))
				accessControlRequestHeaders := requestHeader.Get("Access-Control-Request-Headers")

				for _, corsEndpointSpecification := range corsEndpointSpecifications {
					method := strings.ToUpper(corsEndpointSpecification.Method)

					corsConfiguration.Methods = append(corsConfiguration.Methods, strings.ToUpper(method))
					if method == http.MethodGet {
						corsConfiguration.Methods = append(corsConfiguration.Methods, http.MethodHead)
					}

					if method != accessControlRequestMethod {
						continue
					}

					corsRequestParser := corsEndpointSpecification.CorsRequestParser
					// Sanity check. Should not be `nil` based on the previous check.
					if utils.IsNil(corsRequestParser) {
						return nil, &muxTypesResponseError.ResponseError{
							ServerError: motmedelErrors.NewWithTrace(
								fmt.Errorf("%w (cors)", muxErrors.ErrNilRequestParser),
							),
						}
					}

					endpointCorsConfiguration, responseError := corsRequestParser.Parse(request)
					if responseError != nil {
						return nil, responseError
					}
					if endpointCorsConfiguration == nil {
						continue
					}

					corsConfiguration.Origin = endpointCorsConfiguration.Origin
					corsConfiguration.Headers = endpointCorsConfiguration.Headers
					corsConfiguration.Credentials = endpointCorsConfiguration.Credentials
					corsConfiguration.MaxAge = endpointCorsConfiguration.MaxAge
				}

				if origin := corsConfiguration.Origin; origin != "" {
					headerEntries = append(
						headerEntries,
						&muxTypesResponse.HeaderEntry{Name: "Access-Control-Allow-Origin", Value: origin},
					)
				}

				if methods := corsConfiguration.Methods; len(methods) > 0 {
					uniqueMethods := motmedelIter.Set(methods)
					slices.Sort(uniqueMethods)

					headerEntries = append(
						headerEntries,
						&muxTypesResponse.HeaderEntry{Name: "Access-Control-Allow-Methods", Value: strings.Join(uniqueMethods, ", ")},
					)
				}

				if headers := corsConfiguration.Headers; len(headers) > 0 && accessControlRequestHeaders != "" {
					headerEntries = append(
						headerEntries,
						&muxTypesResponse.HeaderEntry{Name: "Access-Control-Allow-Headers", Value: strings.Join(headers, ", ")},
					)
				}

				if credentials := corsConfiguration.Credentials; credentials {
					headerEntries = append(
						headerEntries,
						&muxTypesResponse.HeaderEntry{Name: "Access-Control-Allow-Credentials", Value: "true"},
					)
				}

				if maxAge := corsConfiguration.MaxAge; maxAge > 0 {
					headerEntries = append(
						headerEntries,
						&muxTypesResponse.HeaderEntry{Name: "Access-Control-Max-Age", Value: fmt.Sprintf("%d", maxAge)},
					)
				}
			}

			return &muxTypesResponse.Response{Headers: headerEntries}, nil
		}

		return nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusMethodNotAllowed,
				fmt.Sprintf("Expected %s.", expectedMethodsString),
				nil,
			),
			Headers: headerEntries,
		}
	}

	var corsHeaderEntries []*muxTypesResponse.HeaderEntry
	if corsRequestParser := endpointSpecification.CorsRequestParser; !utils.IsNil(corsRequestParser) {
		corsConfiguration, responseError := corsRequestParser.Parse(request)
		if responseError != nil {
			return nil, responseError
		}

		if corsConfiguration != nil {
			if origin := corsConfiguration.Origin; origin != "" {
				corsHeaderEntries = append(
					corsHeaderEntries,
					&muxTypesResponse.HeaderEntry{Name: "Access-Control-Allow-Origin", Value: origin},
				)
			}

			if credentials := corsConfiguration.Credentials; credentials {
				corsHeaderEntries = append(
					corsHeaderEntries,
					&muxTypesResponse.HeaderEntry{Name: "Access-Control-Allow-Credentials", Value: "true"},
				)
			}

			if exposeHeaders := corsConfiguration.ExposeHeaders; len(exposeHeaders) > 0 {
				corsHeaderEntries = append(
					corsHeaderEntries,
					&muxTypesResponse.HeaderEntry{
						Name:  "Access-Control-Expose-Headers",
						Value: strings.Join(exposeHeaders, ", "),
					},
				)
			}
		}
	}

	// Perform rate limiting, if specified.

	if rateLimitingConfiguration := endpointSpecification.RateLimitingConfiguration; rateLimitingConfiguration != nil {
		if responseError := muxInternalMux.HandleRateLimiting(rateLimitingConfiguration, request); responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
	}

	// Examine fetch metadata

	if !endpointSpecification.DisableFetchMedata {
		if responseError := muxInternalMux.HandleFetchMetadata(requestHeader, request.Method); responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
	}

	// Check authentication.

	if configuration := endpointSpecification.AuthenticationConfiguration; configuration != nil {
		parser := configuration.Parser
		// Special case. For the authentication configuration, not having a parser is an error.
		if utils.IsNil(parser) {
			return nil, &muxTypesResponseError.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(fmt.Errorf("%w (authentication)", muxErrors.ErrNilRequestParser)),
				Headers:     corsHeaderEntries,
			}
		}

		parsedAuthentication, responseError := parser.Parse(request)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}

		request = request.WithContext(
			context.WithValue(request.Context(), parsing.ParsedRequestAuthenticationContextKey, parsedAuthentication),
		)
	}

	// Obtain the parsed url.

	if configuration := endpointSpecification.UrlParserConfiguration; configuration != nil {
		if parser := configuration.Parser; !utils.IsNil(parser) {
			parsedUrl, responseError := parser.Parse(request)
			if responseError != nil {
				responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
				return nil, responseError
			}

			request = request.WithContext(
				context.WithValue(request.Context(), parsing.ParsedRequestUrlContextKey, parsedUrl),
			)
		}
	}

	// Obtain the parsed header.

	if configuration := endpointSpecification.HeaderParserConfiguration; configuration != nil {
		if parser := configuration.Parser; !utils.IsNil(parser) {
			parsedHeader, responseError := parser.Parse(request)
			if responseError != nil {
				responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
				return nil, responseError
			}

			request = request.WithContext(
				context.WithValue(request.Context(), parsing.ParsedRequestHeaderContextKey, parsedHeader),
			)
		}
	}

	// Validate body parameters and obtain and validate the body

	emptyOption := parsing.BodyOptional
	var expectedContentType string
	var bodyParser body_parser.BodyParser[any]
	var maxBytes int64

	if request.Method == http.MethodGet || request.Method == http.MethodHead || request.Method == http.MethodTrace {
		emptyOption = parsing.BodyForbidden
	}

	// Obtain validation options from the handler specification configuration.
	if bodyParserConfiguration := endpointSpecification.BodyParserConfiguration; bodyParserConfiguration != nil {
		emptyOption = bodyParserConfiguration.EmptyOption
		expectedContentType = bodyParserConfiguration.ContentType
		bodyParser = bodyParserConfiguration.Parser
		maxBytes = bodyParserConfiguration.MaxBytes
	}

	// Validate Content-Type (parse and match header value against accepted value)
	if expectedContentType != "" {
		if responseError = muxInternalMux.ValidateContentType(expectedContentType, requestHeader); responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
	}

	if emptyOption == parsing.BodyForbidden {
		request.Body = http.MaxBytesReader(responseWriter, request.Body, 0)
	} else if maxBytes > 0 {
		request.Body = http.MaxBytesReader(responseWriter, request.Body, maxBytes)
	}

	allowEmptyBody := emptyOption == parsing.BodyOptional || emptyOption == parsing.BodyForbidden

	// Validate Content-Length (parse and check if empty is accepted)
	if responseError := muxInternalMux.ValidateContentLength(allowEmptyBody, requestHeader); responseError != nil {
		responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
		return nil, responseError
	}

	// Obtain the request body
	requestBody, responseError := muxInternalMux.ObtainRequestBody(
		request.Context(),
		request.ContentLength,
		request.Body,
		maxBytes,
	)
	if responseError != nil {
		responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
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
			Headers: corsHeaderEntries,
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
				Headers: corsHeaderEntries,
			}
		}
	}

	// Parse the body.
	if !utils.IsNil(bodyParser) {
		parsedBody, responseError := bodyParser.Parse(request, requestBody)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}

		request = request.WithContext(
			context.WithValue(request.Context(), parsing.ParsedRequestBodyContextKey, parsedBody),
		)
	}

	// Obtain a response

	var handlerResponseHeaders []*muxTypesResponse.HeaderEntry
	var response *muxTypesResponse.Response

	// Respond with dynamic content via a handler.
	handler := endpointSpecification.Handler
	if handler != nil {
		response, responseError = handler(request, requestBody)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
		if response != nil {
			handlerResponseHeaders = response.Headers
		}
	}

	// Respond with static content.
	staticContent := endpointSpecification.StaticContent
	if staticContent != nil {
		var isCached bool
		isCached, responseError = muxInternalMux.ObtainIsCached(staticContent, requestHeader)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}

		var acceptEncoding *motmedelHttpTypes.AcceptEncoding
		contentNegotiation, _ := request.Context().Value(muxContext.ContentNegotiationContextKey).(*motmedelHttpTypes.ContentNegotiation)
		if contentNegotiation != nil {
			acceptEncoding = contentNegotiation.AcceptEncoding
		}

		response, responseError = muxInternalMux.ObtainStaticContentResponse(
			staticContent,
			isCached,
			requestHeader,
			acceptEncoding,
		)
	}

	if responseError != nil {
		responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
		return nil, responseError
	}

	if response == nil {
		response = &muxTypesResponse.Response{}
	}

	// If both a handler and static content are specified, the handler response headers are added to the static content
	// response headers.
	if handler != nil && staticContent != nil {
		response.Headers = append(response.Headers, handlerResponseHeaders...)
	}

	if !endpointSpecification.DisableFetchMedata {
		response.Headers = append(
			response.Headers,
			&muxTypesResponse.HeaderEntry{
				Name:  "Vary",
				Value: "Sec-Fetch-Dest, Sec-Fetch-Mode, Sec-Fetch-Site",
			},
		)
	}

	response.Headers = append(response.Headers, corsHeaderEntries...)
	return response, nil
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

	endpointSpecificationMap := mux.EndpointSpecificationMap
	if endpointSpecificationMap == nil {
		endpointSpecificationMap = make(map[string]map[string]*muxTypesEnpointSpecification.EndpointSpecification)
	}

	for _, specification := range specifications {
		if specification == nil {
			continue
		}
		methodToEndpointSpecification, ok := endpointSpecificationMap[specification.Path]
		if !ok {
			methodToEndpointSpecification = make(map[string]*muxTypesEnpointSpecification.EndpointSpecification)
			endpointSpecificationMap[specification.Path] = methodToEndpointSpecification
		}

		methodToEndpointSpecification[strings.ToUpper(specification.Method)] = specification
	}

	mux.EndpointSpecificationMap = endpointSpecificationMap
}

func (mux *Mux) Delete(specifications ...*muxTypesEnpointSpecification.EndpointSpecification) {
	if len(specifications) == 0 {
		return
	}

	endpointSpecificationMap := mux.EndpointSpecificationMap
	if endpointSpecificationMap == nil {
		return
	}

	for _, specification := range specifications {
		methodToEndpointSpecification, ok := endpointSpecificationMap[specification.Path]
		if !ok {
			return
		}

		delete(methodToEndpointSpecification, strings.ToUpper(specification.Method))

		if len(methodToEndpointSpecification) == 0 {
			delete(endpointSpecificationMap, specification.Path)
		}
	}
}

func (mux *Mux) Get(path string, method string) *muxTypesEnpointSpecification.EndpointSpecification {
	endpointSpecificationMap := mux.EndpointSpecificationMap
	if endpointSpecificationMap == nil {
		return nil
	}

	methodToEndpointSpecification, ok := endpointSpecificationMap[path]
	if !ok || methodToEndpointSpecification == nil {
		return nil
	}

	return methodToEndpointSpecification[strings.ToUpper(method)]
}

func (mux *Mux) GetDocumentEndpointSpecifications() []*muxTypesEnpointSpecification.EndpointSpecification {
	var specifications []*muxTypesEnpointSpecification.EndpointSpecification

	for _, methodMap := range mux.EndpointSpecificationMap {
		for _, specification := range methodMap {
			staticContent := specification.StaticContent
			if staticContent == nil {
				continue
			}

			var isDocument bool
			for _, header := range staticContent.Headers {
				if header == nil {
					continue
				}

				if strings.ToLower(header.Name) == "content-type" && strings.ToLower(header.Value) == "text/html" {
					isDocument = true
					break
				}
			}

			if !isDocument {
				continue
			}

			specifications = append(specifications, specification)
		}
	}

	return specifications
}

func (mux *Mux) DuplicateEndpointSpecification(endpointSpecification *muxTypesEnpointSpecification.EndpointSpecification, routes ...string) error {
	if endpointSpecification == nil {
		return motmedelErrors.NewWithTrace(muxErrors.ErrNilEndpointSpecification)
	}

	for _, route := range routes {
		specification := *endpointSpecification
		specification.Path = route

		mux.Add(&specification)
	}

	return nil
}
