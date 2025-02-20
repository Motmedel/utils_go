package mux

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypes "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	muxTypesRateLimiting "github.com/Motmedel/utils_go/pkg/http/mux/types/rate_limiting"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	muxUtilsCaching "github.com/Motmedel/utils_go/pkg/http/mux/utils/caching"
	muxUtilsContentEncoding "github.com/Motmedel/utils_go/pkg/http/mux/utils/content_encoding"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/accept_encoding"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"io"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
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
			ServerError: motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	getKeyFunc := rateLimitingConfiguration.GetKey
	if getKeyFunc == nil {
		getKeyFunc = muxTypesRateLimiting.DefaultGetRateLimitingKey
	}

	key, err := getKeyFunc(request)
	if err != nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("get key func: %w", err)),
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
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(http.StatusTooManyRequests, "", nil),
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

func ValidateContentType(expectedContentType string, requestHeader http.Header) *muxTypesResponseError.ResponseError {
	// TODO: Error case?
	if expectedContentType == "" {
		return nil
	}

	if requestHeader == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	// TODO: Am I Sure it is accept and now `Allow`?
	acceptedContentTypeHeaders := []*muxTypesResponse.HeaderEntry{{Name: "Accept", Value: expectedContentType}}

	if _, ok := requestHeader["Content-Type"]; !ok {
		return &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnsupportedMediaType,
				"Missing Content-Type.",
				nil,
			),
			Headers: acceptedContentTypeHeaders,
		}
	}

	contentTypeData := []byte(requestHeader.Get("Content-Type"))
	contentType, err := content_type.ParseContentType(contentTypeData)
	if err != nil {
		wrappedErr := motmedelErrors.MakeError(fmt.Errorf("parse content type: %w", err), contentTypeData)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrSyntaxError, motmedelErrors.ErrSemanticError) {
			return &muxTypesResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusBadRequest,
					"Malformed Content-Type.",
					nil,
				),
			}
		}
		return &muxTypesResponseError.ResponseError{ServerError: wrappedErr}
	}
	if contentType == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilContentType),
		}
	}

	// TODO: The specification could require a certain charset too?
	fullNormalizeContentTypeString := contentType.GetFullType(true)
	if fullNormalizeContentTypeString != expectedContentType {
		return &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnsupportedMediaType,
				fmt.Sprintf(
					"Expected Content-Type to be \"%s\", observed \"%s\"",
					expectedContentType,
					fullNormalizeContentTypeString,
				),
				nil,
			),
			Headers: acceptedContentTypeHeaders,
		}
	}

	return nil
}

func ValidateContentLength(allowEmpty bool, requestHeader http.Header) *muxTypesResponseError.ResponseError {
	if requestHeader == nil {
		return &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
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
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusBadRequest,
					"Malformed Content-Length.",
					nil,
				),
				ClientError: motmedelErrors.MakeErrorWithStackTrace(
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
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				zeroContentLengthStatusCode,
				zeroContentLengthMessage,
				nil,
			),
		}
	}

	return nil
}

func ObtainRequestBody(contentLength int64, bodyReader io.ReadCloser) ([]byte, *muxTypesResponseError.ResponseError) {
	if bodyReader == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilHttpRequestBodyReader),
		}
	}

	if contentLength >= 0 {
		var err error
		requestBody, err := io.ReadAll(bodyReader)
		if err != nil {
			return nil, &muxTypesResponseError.ResponseError{
				ServerError: motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("io read all request body: %w", err)),
			}
		}
		defer bodyReader.Close()

		return requestBody, nil
	}

	return nil, nil
}

func ObtainEndpointSpecification(
	endpointSpecificationMap map[string]map[string]*muxTypes.EndpointSpecification,
	request *http.Request,
) (*muxTypes.EndpointSpecification, *muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
	if len(endpointSpecificationMap) == 0 {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(http.StatusNotFound, "", nil),
		}
	}

	if request == nil {
		return nil, nil, &muxTypesResponseError.ResponseError{ServerError: motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequest)}
	}

	requestMethod := strings.ToUpper(request.Method)

	effectiveLookupMethod := requestMethod
	if requestMethod == http.MethodHead {
		// A HEAD request is to be processed as if it were a GET request. But signal not to write a body.
		effectiveLookupMethod = http.MethodGet
	}

	requestUrl := request.URL
	if requestUrl == nil {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(http.StatusNotFound, "", nil),
		}
	}

	methodToHandlerSpecification, ok := endpointSpecificationMap[requestUrl.Path]
	if !ok {
		return nil, nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(http.StatusNotFound, "", nil),
		}
	}

	handlerSpecification, ok := methodToHandlerSpecification[effectiveLookupMethod]
	if !ok {
		allowedMethods := slices.Collect(maps.Keys(methodToHandlerSpecification))

		if _, ok := methodToHandlerSpecification[http.MethodOptions]; !ok {
			allowedMethods = append(allowedMethods, http.MethodOptions)
		}

		if _, ok := methodToHandlerSpecification[http.MethodHead]; !ok {
			if _, ok := methodToHandlerSpecification[http.MethodGet]; ok {
				allowedMethods = append(allowedMethods, http.MethodHead)
			}
		}

		expectedMethodsString := strings.Join(allowedMethods, ", ")
		headerEntries := []*muxTypesResponse.HeaderEntry{{Name: "Allow", Value: expectedMethodsString}}

		if effectiveLookupMethod == http.MethodOptions {
			return nil, &muxTypesResponse.Response{Headers: headerEntries}, nil
		}

		return nil, nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusMethodNotAllowed,
				fmt.Sprintf("Expected \"%s\", observed \"%s\"", expectedMethodsString, requestMethod),
				nil,
			),
			Headers: headerEntries,
		}
	}

	return handlerSpecification, nil, nil
}

func ObtainIsCached(staticContent *muxTypesStaticContent.StaticContent, requestHeader http.Header) (bool, *muxTypesResponseError.ResponseError) {
	if staticContent == nil {
		return false, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilStaticContent),
		}
	}

	if requestHeader == nil {
		return false, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	isCached := muxUtilsCaching.IfNoneMatchCacheHit(requestHeader.Get("If-None-Match"), staticContent.Etag)
	if !isCached {
		var err error
		ifModifiedSince := requestHeader.Get("If-Modified-Since")
		lastModified := staticContent.LastModified
		isCached, err = muxUtilsCaching.IfModifiedSinceCacheHit(ifModifiedSince, lastModified)
		if err != nil {
			wrappedErr := motmedelErrors.MakeError(
				fmt.Errorf("if modified since cache hit: %w", err),
				ifModifiedSince,
				lastModified,
			)
			if errors.Is(err, muxErrors.ErrBadIfModifiedSinceTimestamp) {
				return false, &muxTypesResponseError.ResponseError{
					ProblemDetail: problem_detail.MakeBadRequestProblemDetail("Bad If-Modified-Since value", nil),
					ClientError:   wrappedErr,
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
) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
	if staticContent == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilStaticContent),
		}
	}

	if requestHeader == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.MakeErrorWithStackTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	response := &muxTypesResponse.Response{Headers: staticContent.Headers}
	if isCached {
		response.StatusCode = http.StatusNotModified
	} else {
		encoding := muxUtilsContentEncoding.AcceptContentIdentityIdentifier

		var supportedEncodings []string

		if _, ok := requestHeader["Accept-Encoding"]; ok {
			acceptEncodingData := []byte(requestHeader.Get("Accept-Encoding"))
			acceptEncoding, err := accept_encoding.ParseAcceptEncoding(acceptEncodingData)
			if err != nil {
				wrappedErr := motmedelErrors.MakeError(
					fmt.Errorf("parse accept encoding: %w", err),
					acceptEncodingData,
				)
				if motmedelErrors.IsAny(err, motmedelErrors.ErrSyntaxError, motmedelErrors.ErrSemanticError) {
					return nil, &muxTypesResponseError.ResponseError{
						ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
							http.StatusBadRequest,
							"Malformed Accept-Encoding.",
							nil,
						),
						ClientError: wrappedErr,
					}
				}
				return nil, &muxTypesResponseError.ResponseError{ServerError: wrappedErr}
			}
			if acceptEncoding == nil {
				return nil, &muxTypesResponseError.ResponseError{
					ServerError: motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilAcceptEncoding),
				}
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

			encoding = muxUtilsContentEncoding.GetMatchingContentEncoding(
				acceptEncoding.GetPriorityOrderedEncodings(),
				supportedEncodings,
			)
		}

		if encoding == "" {
			return nil, &muxTypesResponseError.ResponseError{
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnsupportedMediaType,
					"No content encoding could be negotiated.",
					nil,
				),
				Headers: []*muxTypesResponse.HeaderEntry{
					{Name: "Accept-Encoding", Value: strings.Join(supportedEncodings, ", ")},
				},
			}
		}

		response.StatusCode = http.StatusOK
		if encoding == muxUtilsContentEncoding.AcceptContentIdentityIdentifier {
			response.Body = staticContent.Data
		} else {
			response.Headers = append(
				response.Headers,
				&muxTypesResponse.HeaderEntry{Name: "Content-Encoding", Value: encoding},
			)

			contentEncodingToData := staticContent.ContentEncodingToData
			if contentEncodingToData == nil {
				return nil, &muxTypesResponseError.ResponseError{
					ServerError: motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilContentEncodingToData),
				}
			}

			response.Body = contentEncodingToData[encoding].Data
		}
	}

	return response, nil
}
