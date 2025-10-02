package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelTlsTypes "github.com/Motmedel/utils_go/pkg/tls/types"
)

const AcceptContentIdentity = "identity"

func DefaultRetryResponseChecker(response *http.Response, err error) bool {
	if response != nil {
		return response.StatusCode == 429 || response.StatusCode >= 500
	}

	if err != nil {
		return true
	}

	return false
}

func getRetryAfterTime(retryAfterValue string, referenceTime *time.Time) *time.Time {
	if retryAfterValue == "" {
		return nil
	}

	if retryAfterTimestamp, err := time.Parse(time.RFC1123, retryAfterValue); err != nil {
		return &retryAfterTimestamp
	}

	if retryAfterDelay, err := strconv.Atoi(retryAfterValue); err == nil {
		if referenceTime == nil {
			t := time.Now()
			referenceTime = &t
		}

		// Add one more second for rounding.
		waitTime := referenceTime.Add(time.Duration(retryAfterDelay+1) * time.Second)
		return &waitTime
	}

	return nil
}

func fetch(
	ctx context.Context,
	request *http.Request,
	httpClient *http.Client,
	options *motmedelHttpTypes.FetchOptions,
) (*http.Response, []byte, error) {
	if request == nil {
		return nil, nil, nil
	}

	if httpClient == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext)
	if !ok || httpContext == nil {
		httpContext = &motmedelHttpTypes.HttpContext{}
	}
	ctxWithHttpContext := motmedelHttpContext.WithHttpContextValue(context.Background(), httpContext)

	httpContext.Request = request
	if options != nil {
		httpContext.RequestBody = options.Body
	}

	response, err := httpClient.Do(request)
	httpContext.Response = response
	if err != nil {
		return nil, nil, motmedelErrors.NewWithTraceCtx(
			ctxWithHttpContext,
			fmt.Errorf("http client do: %w", err),
		)
	}
	if response == nil {
		return nil, nil, motmedelErrors.NewWithTraceCtx(ctxWithHttpContext, motmedelHttpErrors.ErrNilHttpResponse)
	}
	responseBody := response.Body
	if responseBody == nil {
		return nil, nil, motmedelErrors.NewWithTraceCtx(
			ctxWithHttpContext,
			motmedelHttpErrors.ErrNilHttpResponseBodyReader,
		)
	}

	var skipReadResponseBody bool
	if options != nil {
		skipReadResponseBody = options.SkipReadResponseBody
	}

	var responseBodyData []byte

	if !skipReadResponseBody {
		responseBodyData, err = io.ReadAll(responseBody)
		defer func() {
			if err := responseBody.Close(); err != nil {
				slog.Warn(fmt.Sprintf("close response body: %v", err))
			}
		}()

		if err != nil {
			return response, nil, motmedelErrors.NewWithTraceCtx(
				ctxWithHttpContext,
				fmt.Errorf("io read all (response body): %w", err),
			)
		}

		httpContext.ResponseBody = responseBodyData
	}

	if responseTls := response.TLS; responseTls != nil {
		tlsContext := httpContext.TlsContext
		if tlsContext == nil {
			tlsContext = &motmedelTlsTypes.TlsContext{}
			httpContext.TlsContext = tlsContext
		}

		tlsContext.ConnectionState = responseTls
		tlsContext.ClientInitiated = true
	}

	var skipErrorOnStatus bool
	if options != nil {
		skipErrorOnStatus = options.SkipErrorOnStatus
	}

	if !skipErrorOnStatus {
		if !strings.HasPrefix(strconv.Itoa(response.StatusCode), "2") {
			return response, responseBodyData, motmedelErrors.NewWithTraceCtx(
				ctxWithHttpContext,
				&motmedelHttpErrors.Non2xxStatusCodeError{StatusCode: response.StatusCode},
			)
		}
	}

	return response, responseBodyData, nil
}

func fetchWithRetryConfig(
	ctx context.Context,
	request *http.Request,
	httpClient *http.Client,
	options *motmedelHttpTypes.FetchOptions,
) (*http.Response, []byte, error) {
	if request == nil {
		return nil, nil, nil
	}

	if httpClient == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	var retryConfiguration *motmedelHttpTypes.RetryConfiguration
	if options != nil {
		retryConfiguration = options.RetryConfig
	}
	if retryConfiguration == nil {
		retryConfiguration = ctx.Value(motmedelHttpContext.RetryConfigurationContextKey).(*motmedelHttpTypes.RetryConfiguration)
	}

	var retryCount int
	var baseDelay time.Duration
	var maximumWaitTime time.Duration
	var checkRetryResponse func(*http.Response, error) bool
	if retryConfiguration != nil {
		retryCount = max(retryCount, retryConfiguration.Count)
		baseDelay = retryConfiguration.BaseDelay
		maximumWaitTime = retryConfiguration.MaximumWaitTime
		checkRetryResponse = retryConfiguration.CheckResponse
	}

	var err error
	var response *http.Response
	var responseBody []byte

	// TODO: Do something with http context and extra

	for i := 0; i < (1 + retryCount); i++ {
		if i != 0 {
			// Wait before the next response.

			var waitUntil *time.Time

			// Use information from the previous response to ascertain a waiting time.
			if response != nil {
				responseHeader := response.Header

				if responseHeader != nil {
					// Use the response header Date as the reference time for Retry-After delay values if available,
					//otherwise the current time.
					referenceTime := time.Now()
					if responseDate, err := time.Parse(time.RFC1123, responseHeader.Get("Date")); err != nil {
						referenceTime = responseDate
					}
					waitUntil = getRetryAfterTime(responseHeader.Get("Retry-After"), &referenceTime)
					if waitUntil != nil && maximumWaitTime != 0 {
						if time.Until(*waitUntil) > maximumWaitTime {
							break
						}
					}
				}
			}

			// If the response provided no information about waiting time, use exponential back-off.
			if waitUntil == nil {
				if baseDelay == 0 {
					// If no base delay was provided, the default is 500 ms.
					baseDelay = time.Duration(500) * time.Millisecond
				}
				// baseDelay * 2^(i-1)
				waitDuration := baseDelay * (1 << (i - 1))
				if maximumWaitTime != 0 {
					// Don't let the calculated wait time exceed the maximum wait time.
					waitDuration = min(waitDuration, maximumWaitTime)
				}

				t := time.Now().Add(waitDuration)
				waitUntil = &t
			}

			// TODO: Add jitter?

			duration := time.Until(*waitUntil)
			if duration > 0 {
				time.Sleep(duration)
			}
		}

		response, responseBody, err = fetch(ctx, request, httpClient, options)

		if checkRetryResponse == nil {
			checkRetryResponse = DefaultRetryResponseChecker
		}
		if !checkRetryResponse(response, err) {
			break
		}
		if err != nil && i != 0 {
			err = &motmedelHttpErrors.ReattemptFailedError{Cause: err, Attempt: i + 1}
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("fetch: %w", err)
	}

	return response, responseBody, nil
}

func FetchWithRequest(
	ctx context.Context,
	request *http.Request,
	httpClient *http.Client,
	options *motmedelHttpTypes.FetchOptions,
) (*http.Response, []byte, error) {
	if request == nil {
		return nil, nil, nil
	}

	if httpClient == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	var retryConfiguration *motmedelHttpTypes.RetryConfiguration

	if options != nil {
		retryConfiguration = options.RetryConfig
	}

	if retryConfiguration == nil {
		retryConfiguration, _ = ctx.Value(motmedelHttpContext.RetryConfigurationContextKey).(*motmedelHttpTypes.RetryConfiguration)
	}

	var response *http.Response
	var responseBody []byte
	var err error
	var errString string

	if retryConfiguration != nil {
		response, responseBody, err = fetchWithRetryConfig(ctx, request, httpClient, options)
		errString = " with retry config"
	} else {
		response, responseBody, err = fetch(ctx, request, httpClient, options)
	}
	if err != nil {
		return response, responseBody, fmt.Errorf("fetch%s: %w", errString, err)
	}

	return response, responseBody, nil
}

func Fetch(
	ctx context.Context,
	url string,
	httpClient *http.Client,
	options *motmedelHttpTypes.FetchOptions,
) (*http.Response, []byte, error) {
	if url == "" {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyUrl)
	}

	if httpClient == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	var method string
	var body []byte
	var headers map[string]string
	if options != nil {
		method = options.Method
		headers = options.Headers
		body = options.Body
	}

	if method == "" {
		method = http.MethodGet
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, motmedelErrors.NewWithTrace(fmt.Errorf("http new request: %w", err), method)
	}
	if request == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest)
	}

	requestHeader := request.Header
	if requestHeader == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader)
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	return FetchWithRequest(ctx, request, httpClient, options)
}

func FetchJson[U any](
	ctx context.Context,
	url string,
	httpClient *http.Client,
	options *motmedelHttpTypes.FetchOptions,
) (*http.Response, *U, error) {
	if url == "" {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyUrl)
	}

	if httpClient == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	if options == nil {
		options = &motmedelHttpTypes.FetchOptions{}
	}

	if options.Headers == nil {
		options.Headers = make(map[string]string)
	}

	optionsHeaders := options.Headers

	if _, ok := optionsHeaders["Content-Type"]; !ok && len(options.Body) > 0 {
		optionsHeaders["Content-Type"] = "application/json"
	}

	if _, ok := optionsHeaders["Accept"]; !ok {
		optionsHeaders["Accept"] = "application/json"
	}

	response, responseBody, err := Fetch(ctx, url, httpClient, options)
	if err != nil {
		return response, nil, motmedelErrors.New(fmt.Errorf("fetch: %w", err), url, httpClient, options)
	}
	if len(responseBody) == 0 {
		return response, nil, nil
	}

	var responseValue *U
	if err = json.Unmarshal(responseBody, &responseValue); err != nil {
		return response, nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("json unmarshal (response body): %w", err),
			responseBody,
		)
	}

	return response, responseValue, nil
}

func FetchJsonWithBody[U any, T any](
	ctx context.Context,
	url string,
	httpClient *http.Client,
	bodyValue *T,
	options *motmedelHttpTypes.FetchOptions,
) (*http.Response, *U, error) {
	if url == "" {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyUrl)
	}

	if httpClient == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpClient)
	}

	var requestBody []byte
	if bodyValue != nil {
		var err error
		requestBody, err = json.Marshal(bodyValue)
		if err != nil {
			return nil, nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("json marshal (body value): %w", err),
				bodyValue,
			)
		}

		if options == nil {
			options = &motmedelHttpTypes.FetchOptions{}
		}

		options.Body = requestBody
	}

	return FetchJson[U](ctx, url, httpClient, options)
}

func GetMatchingContentEncoding(
	clientSupportedEncodings []*motmedelHttpTypes.Encoding,
	serverSupportedEncodingIdentifiers []string,
) string {
	if len(clientSupportedEncodings) == 0 {
		return AcceptContentIdentity
	}

	disallowIdentity := false

	for _, clientEncoding := range clientSupportedEncodings {
		coding := strings.ToLower(clientEncoding.Coding)
		qualityValue := clientEncoding.QualityValue

		if coding == "*" {
			if qualityValue == 0 {
				disallowIdentity = true
			} else {
				if len(serverSupportedEncodingIdentifiers) != 0 {
					return serverSupportedEncodingIdentifiers[0]
				} else {
					if !disallowIdentity {
						return AcceptContentIdentity
					}
				}
			}
		}

		if coding == AcceptContentIdentity {
			if qualityValue == 0 {
				disallowIdentity = true
			} else {
				return AcceptContentIdentity
			}
		}

		if qualityValue == 0 {
			continue
		}

		for _, supportedEncoding := range serverSupportedEncodingIdentifiers {
			if clientEncoding.Coding == supportedEncoding {
				return supportedEncoding
			}
		}
	}

	if !disallowIdentity {
		return AcceptContentIdentity
	} else {
		return ""
	}
}

func GetMatchingAccept(
	clientSupportedMediaRanges []*motmedelHttpTypes.MediaRange,
	serverSupportedMediaRanges []*motmedelHttpTypes.ServerMediaRange,
) *motmedelHttpTypes.ServerMediaRange {
	if len(clientSupportedMediaRanges) == 0 || len(serverSupportedMediaRanges) == 0 {
		return nil
	}

	for _, clientMediaRange := range clientSupportedMediaRanges {
		if clientMediaRange == nil {
			continue
		}

		clientType := strings.ToLower(clientMediaRange.Type)
		clientSubtype := strings.ToLower(clientMediaRange.Subtype)

		for _, serverMediaRange := range serverSupportedMediaRanges {
			if serverMediaRange == nil {
				continue
			}

			if clientType == "*" && clientSubtype == "*" {
				return serverMediaRange
			}

			serverType := strings.ToLower(serverMediaRange.Type)
			serverSubtype := strings.ToLower(serverMediaRange.Subtype)

			if (clientType == "*" || clientType == serverType) && (clientSubtype == "*" || clientSubtype == serverSubtype) {
				return serverMediaRange
			}
		}
	}

	return nil
}

func ParseLastModifiedTimestamp(timestamp string) (time.Time, error) {
	if t, err := time.Parse(time.RFC1123, timestamp); err != nil {
		return time.Time{}, motmedelErrors.NewWithTrace(
			fmt.Errorf(
				"%w: time parse rfc1123: %w",
				motmedelHttpErrors.ErrBadIfModifiedSinceTimestamp,
				err,
			),
			timestamp,
		)
	} else {
		return t, nil
	}
}

func IfNoneMatchCacheHit(ifNoneMatchValue string, etag string) bool {
	if ifNoneMatchValue == "" || etag == "" {
		return false
	}

	return ifNoneMatchValue == etag
}

func IfModifiedSinceCacheHit(ifModifiedSinceValue string, lastModifiedValue string) (bool, error) {
	if ifModifiedSinceValue == "" || lastModifiedValue == "" {
		return false, nil
	}

	ifModifiedSinceTimestamp, err := ParseLastModifiedTimestamp(ifModifiedSinceValue)
	if err != nil {
		return false, motmedelErrors.New(
			fmt.Errorf("parse last modified timestamp (If-Modified-Since): %w", err),
			ifModifiedSinceValue,
		)
	}

	lastModifiedTimestamp, err := ParseLastModifiedTimestamp(lastModifiedValue)
	if err != nil {
		return false, motmedelErrors.New(
			fmt.Errorf("parse last modified timestamp (Last-Modified): %w", err),
			lastModifiedValue,
		)
	}

	return ifModifiedSinceTimestamp.Equal(lastModifiedTimestamp) || lastModifiedTimestamp.Before(ifModifiedSinceTimestamp), nil
}

func MakeStrongEtag(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return fmt.Sprintf("\"%x\"", h.Sum(nil))
}

// NOTE: Copied from the standard library.

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
