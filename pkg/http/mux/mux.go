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
	muxInternal "github.com/Motmedel/utils_go/pkg/http/mux/internal"
	muxInternalMux "github.com/Motmedel/utils_go/pkg/http/mux/internal/mux"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_loader/body_setting"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser"
	muxTypesEnpointSpecification "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint"
	muxTypesFirewall "github.com/Motmedel/utils_go/pkg/http/mux/types/firewall"
	muxTypesMiddleware "github.com/Motmedel/utils_go/pkg/http/mux/types/middleware"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponseWriter "github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/userer"
	utils2 "github.com/Motmedel/utils_go/pkg/http/mux/utils"
	muxUtilsContentNegotiation "github.com/Motmedel/utils_go/pkg/http/mux/utils/content_negotiation"
	contentSecurityPolicyParsing "github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_security_policy"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	motmedelIter "github.com/Motmedel/utils_go/pkg/iter"
	"github.com/Motmedel/utils_go/pkg/utils"
	"github.com/google/uuid"
)

const (
	contentSecurityPolicyHeaderName = "Content-Security-Policy"
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
			motmedelContext.WithError(
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
						motmedelContext.WithError(
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
		}

		// Trigger a termination of the connection.
		panic(http.ErrAbortHandler)
	} else if verdict == muxTypesFirewall.VerdictReject {
		if firewallResponseError == nil {
			firewallResponseError = &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.New(http.StatusForbidden),
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
	EndpointSpecificationMap map[string]map[string]*muxTypesEnpointSpecification.Endpoint
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

	endpoint, methodToEndpoint, responseError := muxInternalMux.GetEndpoint(
		mux.EndpointSpecificationMap,
		request,
	)
	if responseError != nil {
		return nil, responseError
	}

	// There exists no endpoint for the given method,
	if endpoint == nil {
		// and for no other methods either, which is an error (as "Not Found" should be produced by `GetEndpoint`)
		if len(methodToEndpoint) == 0 {
			return nil, &muxTypesResponseError.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilEndpointSpecification),
			}
		}

		// Produce an OPTIONS response (list allowed methods and/or CORS configuration).

		var allowedMethods []string
		var corsEndpoints []*muxTypesEnpointSpecification.Endpoint

		for method, otherEndpoint := range methodToEndpoint {
			if otherEndpoint == nil {
				continue
			}

			allowedMethods = append(allowedMethods, method)

			if corsParser := otherEndpoint.CorsParser; !utils.IsNil(corsParser) {
				corsEndpoints = append(corsEndpoints, otherEndpoint)
			}
		}

		if _, ok := methodToEndpoint[http.MethodHead]; !ok {
			if _, ok := methodToEndpoint[http.MethodGet]; ok {
				allowedMethods = append(allowedMethods, http.MethodHead)
			}
		}
		if _, ok := methodToEndpoint[http.MethodOptions]; !ok {
			allowedMethods = append(allowedMethods, http.MethodOptions)
		}
		slices.Sort(allowedMethods)

		expectedMethodsString := strings.Join(allowedMethods, ", ")
		headerEntries := []*muxTypesResponse.HeaderEntry{{Name: "Allow", Value: expectedMethodsString}}

		if strings.ToUpper(request.Method) == http.MethodOptions {
			if len(corsEndpoints) > 0 {
				var corsConfiguration motmedelHttpTypes.CorsConfiguration
				accessControlRequestMethod := strings.ToUpper(requestHeader.Get("Access-Control-Request-Method"))
				accessControlRequestHeaders := requestHeader.Get("Access-Control-Request-Headers")

				for _, corsEndpoint := range corsEndpoints {
					method := strings.ToUpper(corsEndpoint.Method)

					corsConfiguration.Methods = append(corsConfiguration.Methods, strings.ToUpper(method))
					if method == http.MethodGet {
						corsConfiguration.Methods = append(corsConfiguration.Methods, http.MethodHead)
					}

					if method != accessControlRequestMethod {
						continue
					}

					corsParser := corsEndpoint.CorsParser
					// Sanity check. Should not be `nil` based on the previous check.
					if utils.IsNil(corsParser) {
						return nil, &muxTypesResponseError.ResponseError{
							ServerError: motmedelErrors.NewWithTrace(
								fmt.Errorf("%w (cors)", muxErrors.ErrNilRequestParser),
							),
						}
					}

					endpointCorsConfiguration, responseError := corsParser.Parse(request)
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
			ProblemDetail: problem_detail.New(
				http.StatusMethodNotAllowed,
				problem_detail_config.WithDetail(fmt.Sprintf("Expected %s.", expectedMethodsString)),
			),
			Headers: headerEntries,
		}
	}

	var corsHeaderEntries []*muxTypesResponse.HeaderEntry
	if corsParser := endpoint.CorsParser; !utils.IsNil(corsParser) {
		corsConfiguration, responseError := corsParser.Parse(request)
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

	if rateLimitingConfiguration := endpoint.RateLimitingConfiguration; rateLimitingConfiguration != nil {
		if responseError := muxInternalMux.HandleRateLimiting(rateLimitingConfiguration, request); responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
	}

	// Examine fetch metadata

	if !endpoint.DisableFetchMedata {
		if responseError := muxInternalMux.HandleFetchMetadata(requestHeader, request.Method); responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
	}

	// Check authentication.

	if authenticationParser := endpoint.AuthenticationParser; !utils.IsNil(authenticationParser) {
		parsedAuthentication, responseError := authenticationParser.Parse(request)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
		}

		request = request.WithContext(
			context.WithValue(request.Context(), utils2.ParsedRequestAuthenticationContextKey, parsedAuthentication),
		)

		if usererAuthenticationData, ok := parsedAuthentication.(userer.Userer); ok {
			if !utils.IsNil(usererAuthenticationData) {
				httpContext.User = usererAuthenticationData.GetUser()
			}
		}
	}

	// Obtain the parsed url.

	if urlParser := endpoint.UrlParser; !utils.IsNil(urlParser) {
		parsedUrl, responseError := urlParser.Parse(request)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}

		request = request.WithContext(
			context.WithValue(request.Context(), utils2.ParsedRequestUrlContextKey, parsedUrl),
		)
	}

	// Obtain the parsed header.

	if headerParser := endpoint.HeaderParser; !utils.IsNil(headerParser) {
		parsedHeader, responseError := headerParser.Parse(request)
		if responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}

		request = request.WithContext(
			context.WithValue(request.Context(), utils2.ParsedRequestHeaderContextKey, parsedHeader),
		)
	}

	// Validate body parameters and obtain and validate the body

	var expectedContentType string
	var maxBytes int64
	var bodyParser body_parser.BodyParser[any]

	emptyOption := body_setting.Optional
	if request.Method == http.MethodGet || request.Method == http.MethodHead || request.Method == http.MethodTrace || request.Method == http.MethodDelete {
		emptyOption = body_setting.Forbidden
	}

	bodyLoader := endpoint.BodyLoader
	if bodyLoader != nil {
		emptyOption = bodyLoader.Setting
		expectedContentType = bodyLoader.ContentType
		maxBytes = bodyLoader.MaxBytes
		bodyParser = bodyLoader.Parser
	}

	// Validate Content-Type (parse and match header value against accepted value)
	if expectedContentType != "" {
		if responseError = muxInternalMux.ValidateContentType(expectedContentType, requestHeader); responseError != nil {
			responseError.Headers = append(responseError.Headers, corsHeaderEntries...)
			return nil, responseError
		}
	}

	if emptyOption == body_setting.Forbidden {
		request.Body = http.MaxBytesReader(responseWriter, request.Body, 0)
	} else if maxBytes > 0 {
		request.Body = http.MaxBytesReader(responseWriter, request.Body, maxBytes)
	}

	allowEmptyBody := emptyOption == body_setting.Optional || emptyOption == body_setting.Forbidden

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
			ProblemDetail: problem_detail.New(
				http.StatusBadRequest,
				problem_detail_config.WithDetail("A body is expected."),
			),
			Headers: corsHeaderEntries,
		}
	}

	// Basic check to see if the request body conforms to the expected content type.
	switch expectedContentType {
	case "application/json":
		if !json.Valid(requestBody) {
			return nil, &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.New(
					http.StatusBadRequest,
					problem_detail_config.WithDetail("Invalid JSON body."),
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
			context.WithValue(request.Context(), utils2.ParsedRequestBodyContextKey, parsedBody),
		)
	}

	// Obtain a response

	var handlerResponseHeaders []*muxTypesResponse.HeaderEntry
	var response *muxTypesResponse.Response

	// Respond with dynamic content via a handler.
	handler := endpoint.Handler
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
	staticContent := endpoint.StaticContent
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

	if !endpoint.DisableFetchMedata {
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

func (mux *Mux) Add(specifications ...*muxTypesEnpointSpecification.Endpoint) {
	if len(specifications) == 0 {
		return
	}

	endpointSpecificationMap := mux.EndpointSpecificationMap
	if endpointSpecificationMap == nil {
		endpointSpecificationMap = make(map[string]map[string]*muxTypesEnpointSpecification.Endpoint)
	}

	for _, specification := range specifications {
		if specification == nil {
			continue
		}
		methodToEndpointSpecification, ok := endpointSpecificationMap[specification.Path]
		if !ok {
			methodToEndpointSpecification = make(map[string]*muxTypesEnpointSpecification.Endpoint)
			endpointSpecificationMap[specification.Path] = methodToEndpointSpecification
		}

		methodToEndpointSpecification[strings.ToUpper(specification.Method)] = specification
	}

	mux.EndpointSpecificationMap = endpointSpecificationMap
}

func (mux *Mux) Delete(specifications ...*muxTypesEnpointSpecification.Endpoint) {
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

func (mux *Mux) Get(path string, method string) *muxTypesEnpointSpecification.Endpoint {
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

func (mux *Mux) GetDocumentEndpointSpecifications() []*muxTypesEnpointSpecification.Endpoint {
	var specifications []*muxTypesEnpointSpecification.Endpoint

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

func (mux *Mux) DuplicateEndpointSpecification(endpointSpecification *muxTypesEnpointSpecification.Endpoint, routes ...string) error {
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

func (mux *Mux) GetContentSecurityPolicy() (*content_security_policy.ContentSecurityPolicy, error) {
	contentSecurityPolicyString := mux.DefaultDocumentHeaders[contentSecurityPolicyHeaderName]
	if contentSecurityPolicyString == "" {
		return nil, nil
	}

	csp, err := contentSecurityPolicyParsing.Parse([]byte(contentSecurityPolicyString))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return csp, nil
}

func (mux *Mux) SetContentSecurityPolicy(csp *content_security_policy.ContentSecurityPolicy) error {
	defaultDocumentHeaders := mux.DefaultDocumentHeaders
	if defaultDocumentHeaders == nil {
		return motmedelErrors.NewWithTrace(fmt.Errorf("%w (default document headers)", motmedelErrors.ErrNilMap))
	}

	if csp == nil {
		defaultDocumentHeaders[contentSecurityPolicyHeaderName] = ""
	} else {
		defaultDocumentHeaders[contentSecurityPolicyHeaderName] = csp.String()
	}

	return nil
}
