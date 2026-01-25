package mux

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypes "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	muxTypesRateLimiting "github.com/Motmedel/utils_go/pkg/http/mux/types/rate_limiting"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
)

func HandleRateLimiting(
	rateLimitingConfiguration *muxTypesRateLimiting.RateLimitingConfiguration,
	request *http.Request,
) *muxTypesResponseError.ResponseError {
	if rateLimitingConfiguration == nil {
		return nil
	}

	if request == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	getKeyFunc := rateLimitingConfiguration.GetKey
	if getKeyFunc == nil {
		getKeyFunc = muxTypesRateLimiting.DefaultGetRateLimitingKey
	}

	key, err := getKeyFunc(request)
	if err != nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(fmt.Errorf("get key func: %w", err)),
		}
	}

	rateLimitingConfiguration.Lookup.Mutex.Lock()
	if rateLimitingConfiguration.Lookup.Map == nil {
		rateLimitingConfiguration.Lookup.Map = make(map[string]*muxTypesRateLimiting.TimerRateLimiter)
	}
	rateLimitingConfiguration.Lookup.Mutex.Unlock()

	timerRateLimiter, ok := rateLimitingConfiguration.Lookup.Map[key]
	if !ok {
		timerRateLimiter = &muxTypesRateLimiting.TimerRateLimiter{
			RateLimiter: muxTypesRateLimiting.RateLimiter{
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
		return &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.New(http.StatusTooManyRequests),
			Headers: []*muxTypesResponse.HeaderEntry{
				{
					Name:  "Retry-After",
					Value: expirationTime.UTC().Format("Mon, 02 Jan 2006 15:04:05") + " GMT",
				},
			},
		}
	}

	return nil
}

func HandleFetchMetadata(requestHeader http.Header, method string) *muxTypesResponseError.ResponseError {
	if requestHeader == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	if method == "" {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyMethod),
		}
	}

	// NOTE: This check is opinionated; embedding is not allowed. For custom fetch metadata logic, disable this
	// check and implement your own in the firewall configuration e.g., plus add the `Vary` header.

	fetchSite := requestHeader.Get("Sec-Fetch-Site")
	if fetchSite == "" || fetchSite == "same-origin" || fetchSite == "same-site" || fetchSite == "none" {
		return nil
	}

	fetchMode := requestHeader.Get("Sec-Fetch-Mode")
	fetchDest := requestHeader.Get("Sec-Fetch-Dest")
	if fetchMode == "navigate" && fetchDest == "document" && method == http.MethodGet {
		return nil
	}

	return &muxTypesResponseError.ResponseError{
		ProblemDetail: problem_detail.New(
			http.StatusForbidden,
			problem_detail_config.WithDetail("Cross-site request blocked by Fetch-Metadata policy."),
		),
	}
}

func ValidateContentType(expectedContentType string, requestHeader http.Header) *muxTypesResponseError.ResponseError {
	// TODO: Error case?
	if expectedContentType == "" {
		return nil
	}

	if requestHeader == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	acceptedContentTypeHeaders := []*muxTypesResponse.HeaderEntry{{Name: "Accept", Value: expectedContentType}}

	if _, ok := requestHeader["Content-Type"]; !ok {
		return &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.New(
				http.StatusUnsupportedMediaType,
				problem_detail_config.WithDetail("Missing Content-Type."),
			),
			Headers: acceptedContentTypeHeaders,
		}
	}

	contentTypeData := []byte(requestHeader.Get("Content-Type"))
	contentType, err := content_type.Parse(contentTypeData)
	if err != nil {
		wrappedErr := motmedelErrors.New(fmt.Errorf("content type parse: %w", err), contentTypeData)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrSyntaxError, motmedelErrors.ErrSemanticError) {
			return &muxTypesResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.New(
					http.StatusBadRequest,
					problem_detail_config.WithDetail("Malformed Content-Type."),
				),
			}
		}
		return &muxTypesResponseError.ResponseError{ServerError: wrappedErr}
	}
	if contentType == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(content_type.ErrNilContentType),
		}
	}

	// TODO: The specification could require a certain charset too?
	fullNormalizeContentTypeString := contentType.GetFullType(true)
	if fullNormalizeContentTypeString != expectedContentType {
		return &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.New(
				http.StatusUnsupportedMediaType,
				problem_detail_config.WithDetail(
					fmt.Sprintf(
						"Expected Content-Type to be %q, observed %q.",
						expectedContentType,
						fullNormalizeContentTypeString,
					),
				),
			),
			Headers: acceptedContentTypeHeaders,
		}
	}

	return nil
}

func ValidateContentLength(allowEmpty bool, requestHeader http.Header) *muxTypesResponseError.ResponseError {
	if requestHeader == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	zeroContentLengthStatusCode := http.StatusLengthRequired
	zeroContentLengthMessage := "A body is expected; Content-Length must be set."

	var contentLength uint64
	if _, ok := requestHeader["Content-Length"]; ok {
		var err error
		headerValue := requestHeader.Get("Content-Length")
		contentLength, err = strconv.ParseUint(headerValue, 10, 64)
		if err != nil {
			return &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.New(
					http.StatusBadRequest,
					problem_detail_config.WithDetail("Malformed Content-Length."),
				),
				ClientError: motmedelErrors.NewWithTrace(
					fmt.Errorf("strconv parse uint: %w", err),
					headerValue, 10, 64,
				),
			}
		}
		if contentLength == 0 {
			zeroContentLengthStatusCode = http.StatusBadRequest
			zeroContentLengthMessage = "A body is expected; Content-Length cannot be 0."
		}
	}

	if !allowEmpty && contentLength == 0 {
		return &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.New(
				zeroContentLengthStatusCode,
				problem_detail_config.WithDetail(zeroContentLengthMessage),
			),
		}
	}

	return nil
}

func ObtainRequestBody(
	ctx context.Context,
	contentLength int64,
	bodyReader io.ReadCloser,
	maxBytes int64,
) ([]byte, *muxTypesResponseError.ResponseError) {
	if bodyReader == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestBodyReader),
		}
	}

	if contentLength >= 0 {
		if contentLength > 0 && maxBytes > 0 && contentLength > maxBytes {
			return nil, &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.New(
					http.StatusRequestEntityTooLarge,
					problem_detail_config.WithDetail(fmt.Sprintf("Limit: %d bytes", maxBytes)),
				),
			}
		}

		var err error
		requestBody, err := io.ReadAll(bodyReader)
		if err != nil {
			wrappedErr := motmedelErrors.NewWithTrace(
				fmt.Errorf("io read all (request body): %w", err),
				bodyReader,
			)

			// NOTE: "Request Entity Too Large" should always be picked up by the content length check, but add this
			// for completion.
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				return nil, &muxTypesResponseError.ResponseError{
					ClientError: wrappedErr,
					ProblemDetail: problem_detail.New(
						http.StatusRequestEntityTooLarge,
						problem_detail_config.WithDetail(fmt.Sprintf("Limit: %d bytes", maxBytesError.Limit)),
					),
				}
			}

			return nil, &muxTypesResponseError.ResponseError{ServerError: wrappedErr}
		}
		defer func() {
			if err := bodyReader.Close(); err != nil {
				slog.WarnContext(
					motmedelContext.WithError(
						ctx,
						motmedelErrors.NewWithTrace(fmt.Errorf("body reader close: %w", err), bodyReader),
					),
					"An error occurred when closing the request body reader.",
				)
			}
		}()

		return requestBody, nil
	}

	return nil, nil
}

func GetEndpointSpecification(
	endpointSpecificationMap map[string]map[string]*muxTypes.EndpointSpecification,
	request *http.Request,
) (*muxTypes.EndpointSpecification, map[string]*muxTypes.EndpointSpecification, *muxTypesResponseError.ResponseError) {
	if len(endpointSpecificationMap) == 0 {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.New(http.StatusNotFound),
		}
	}

	if request == nil {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestUrl := request.URL
	if requestUrl == nil {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestUrl),
		}
	}

	requestMethod := strings.ToUpper(request.Method)
	effectiveLookupMethod := requestMethod
	if requestMethod == http.MethodHead {
		// A HEAD request is to be processed as if it were a GET request. But signal not to write a body.
		effectiveLookupMethod = http.MethodGet
	}

	methodToEndpointSpecification, ok := endpointSpecificationMap[requestUrl.Path]
	if !ok {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.New(http.StatusNotFound),
		}
	}

	endpointSpecification, ok := methodToEndpointSpecification[effectiveLookupMethod]
	if !ok {
		return nil, methodToEndpointSpecification, nil
	}

	return endpointSpecification, methodToEndpointSpecification, nil
}

func ObtainIsCached(staticContent *muxTypesStaticContent.StaticContent, requestHeader http.Header) (bool, *muxTypesResponseError.ResponseError) {
	if staticContent == nil {
		return false, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilStaticContent),
		}
	}

	if requestHeader == nil {
		return false, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	isCached := motmedelHttpUtils.IfNoneMatchCacheHit(requestHeader.Get("If-None-Match"), staticContent.Etag)
	if !isCached {
		var err error
		ifModifiedSince := requestHeader.Get("If-Modified-Since")
		lastModified := staticContent.LastModified
		isCached, err = motmedelHttpUtils.IfModifiedSinceCacheHit(ifModifiedSince, lastModified)
		if err != nil {
			wrappedErr := motmedelErrors.New(
				fmt.Errorf("if modified since cache hit: %w", err),
				ifModifiedSince,
				lastModified,
			)
			if errors.Is(err, motmedelHttpErrors.ErrBadIfModifiedSinceTimestamp) {
				return false, &muxTypesResponseError.ResponseError{
					ProblemDetail: problem_detail.New(
						http.StatusBadRequest,
						problem_detail_config.WithDetail("Bad If-Modified-Since value"),
					),
					ClientError: wrappedErr,
				}
			} else {
				return false, &muxTypesResponseError.ResponseError{
					ServerError: wrappedErr,
				}
			}
		}
	}

	return isCached, nil
}

func ObtainStaticContentResponse(
	staticContent *muxTypesStaticContent.StaticContent,
	isCached bool,
	requestHeader http.Header,
	acceptEncoding *motmedelHttpTypes.AcceptEncoding,
) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
	if staticContent == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilStaticContent),
		}
	}

	if requestHeader == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	// NOTE: It is up to the user to provide the `Vary` header.
	response := &muxTypesResponse.Response{Headers: staticContent.Headers}
	if isCached {
		response.StatusCode = http.StatusNotModified
	} else {
		encoding := motmedelHttpUtils.AcceptContentIdentity

		if acceptEncoding != nil {
			supportedEncodings := slices.Collect(maps.Keys(staticContent.ContentEncodingToData))
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

			encoding = motmedelHttpUtils.GetMatchingContentEncoding(
				acceptEncoding.GetPriorityOrderedEncodings(),
				supportedEncodings,
			)
		}

		if encoding == "" {
			// NOTE: The problem detail won't appear in the response body because not even `identity` is acceptable;
			//	rather, the problem detail specifies the status code only.
			return nil, &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.New(http.StatusNotAcceptable),
			}
		}

		response.StatusCode = http.StatusOK
		if encoding == motmedelHttpUtils.AcceptContentIdentity {
			response.Body = staticContent.Data
		} else {
			response.Headers = append(
				response.Headers,
				&muxTypesResponse.HeaderEntry{Name: "Content-Encoding", Value: encoding},
			)

			contentEncodingToData := staticContent.ContentEncodingToData
			if contentEncodingToData == nil {
				return nil, &muxTypesResponseError.ResponseError{
					ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilContentEncodingToData),
				}
			}

			staticContentData, ok := contentEncodingToData[encoding]
			if !ok {
				return nil, &muxTypesResponseError.ResponseError{
					ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrContentEncodingToDataNotOk),
				}
			}
			if staticContentData == nil {
				return nil, &muxTypesResponseError.ResponseError{
					ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilStaticContentData),
				}
			}

			response.Body = staticContentData.Data
		}
	}

	return response, nil
}
