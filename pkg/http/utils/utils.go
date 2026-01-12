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
	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config"
	motmedelTlsTypes "github.com/Motmedel/utils_go/pkg/tls/types"
	"github.com/Motmedel/utils_go/pkg/utils"
)

const AcceptContentIdentity = "identity"

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

func fetch(ctx context.Context, request *http.Request, fetchConfig *fetch_config.Config) (*http.Response, []byte, error) {
	if request == nil {
		return nil, nil, nil
	}

	if fetchConfig == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilFetchConfig)
	}

	httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext)
	if !ok || httpContext == nil {
		httpContext = &motmedelHttpTypes.HttpContext{}
	}
	ctxWithHttpContext := motmedelHttpContext.WithHttpContextValue(context.Background(), httpContext)

	httpContext.Request = request
	httpContext.RequestBody = fetchConfig.Body

	response, err := fetchConfig.HttpClient.Do(request)
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

	var responseBodyData []byte
	if !fetchConfig.SkipReadResponseBody {
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

	if !fetchConfig.SkipErrorOnStatus {
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
	fetchConfig *fetch_config.Config,
) (*http.Response, []byte, error) {
	if request == nil {
		return nil, nil, nil
	}

	if fetchConfig == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilFetchConfig)
	}

	retryConfig := fetchConfig.RetryConfig
	if retryConfig == nil {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilFetchRetryConfig)
	}

	maximumWaitTime := retryConfig.MaximumWaitTime

	var err error
	var response *http.Response
	var responseBody []byte

	// TODO: Do something with http context and extra

	for i := 0; i < (1 + retryConfig.Count); i++ {
		if i != 0 {
			// Wait before the next response.

			var waitUntil *time.Time

			// Use information from the previous response to ascertain a waiting time.
			if response != nil {
				responseHeader := response.Header

				if responseHeader != nil {
					// Use the response header Date as the reference time for Retry-After delay values if available,
					// otherwise the current time.
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
				// baseDelay * 2^(i-1)
				waitDuration := retryConfig.BaseDelay * (1 << (i - 1))
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

		response, responseBody, err = fetch(ctx, request, fetchConfig)

		if !retryConfig.ResponseChecker.Check(response, err) {
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

func FetchWithRequest(ctx context.Context, request *http.Request, options ...fetch_config.Option) (*http.Response, []byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context err: %w", err)
	}

	if request == nil {
		return nil, nil, nil
	}

	fetchConfig := fetch_config.New(options...)

	var response *http.Response
	var responseBody []byte
	var err error
	var errString string

	if fetchConfig.RetryConfig != nil {
		response, responseBody, err = fetchWithRetryConfig(ctx, request, fetchConfig)
		errString = " with retry config"
	} else {
		response, responseBody, err = fetch(ctx, request, fetchConfig)
	}
	if err != nil {
		return response, responseBody, fmt.Errorf("fetch%s: %w", errString, err)
	}

	return response, responseBody, nil
}

func Fetch(ctx context.Context, url string, options ...fetch_config.Option) (*http.Response, []byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context err: %w", err)
	}

	if url == "" {
		return nil, nil, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyUrl)
	}

	fetchConfig := fetch_config.New(options...)
	method := fetchConfig.Method

	request, err := http.NewRequest(method, url, bytes.NewBuffer(fetchConfig.Body))
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

	for key, value := range fetchConfig.Headers {
		request.Header.Set(key, value)
	}

	return FetchWithRequest(ctx, request, options...)
}

func FetchJson[U any](ctx context.Context, url string, options ...fetch_config.Option) (*http.Response, U, error) {
	var zero U

	if err := ctx.Err(); err != nil {
		return nil, zero, fmt.Errorf("context err: %w", err)
	}

	if url == "" {
		return nil, zero, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyUrl)
	}

	fetchConfig := fetch_config.New(options...)

	headers := fetchConfig.Headers
	if headers == nil {
		headers = make(map[string]string)
	}

	if _, ok := headers["Content-Type"]; !ok && len(fetchConfig.Body) > 0 {
		headers["Content-Type"] = "application/json"
	}

	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = "application/json"
	}

	options = append(options, fetch_config.WithHeaders(headers))
	response, responseBody, err := Fetch(ctx, url, options...)
	if err != nil {
		return response, zero, fmt.Errorf("fetch: %w", err)
	}
	if len(responseBody) == 0 {
		return response, zero, nil
	}

	var responseValue U
	if err = json.Unmarshal(responseBody, &responseValue); err != nil {
		return response, zero, motmedelErrors.NewWithTrace(
			fmt.Errorf("json unmarshal (response body): %w", err),
			responseBody,
		)
	}

	return response, responseValue, nil
}

func FetchJsonWithBody[U any, T any](ctx context.Context, url string, bodyValue T, options ...fetch_config.Option) (*http.Response, U, error) {
	var zero U

	if err := ctx.Err(); err != nil {
		return nil, zero, fmt.Errorf("context err: %w", err)
	}

	if url == "" {
		return nil, zero, motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyUrl)
	}

	var requestBody []byte
	if !utils.IsNil(bodyValue) {
		var err error
		requestBody, err = json.Marshal(bodyValue)
		if err != nil {
			return nil, zero, motmedelErrors.NewWithTrace(
				fmt.Errorf("json marshal (body value): %w", err),
				bodyValue,
			)
		}

		options = append(options, fetch_config.WithBody(requestBody))
	}

	return FetchJson[U](ctx, url, options...)
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
	}

	return ""
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
	t, err := time.Parse(time.RFC1123, timestamp)

	if err != nil {
		return time.Time{}, motmedelErrors.NewWithTrace(
			fmt.Errorf(
				"%w: time parse rfc1123: %w",
				motmedelHttpErrors.ErrBadIfModifiedSinceTimestamp,
				err,
			),
			timestamp,
		)
	}

	return t, nil
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

func BasicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func GetSingleHeader(name string, header http.Header) (string, error) {
	if header == nil {
		return "", motmedelErrors.NewWithTrace(motmedelErrors.ErrNilMap)
	}

	name = http.CanonicalHeaderKey(name)

	headerValues, ok := header[name]
	if !ok {
		return "", motmedelErrors.NewWithTrace(fmt.Errorf("%w (%s)", motmedelHttpErrors.ErrMissingHeader, name))
	}
	if len(headerValues) != 1 {
		return "", motmedelErrors.NewWithTrace(fmt.Errorf("%w (%s)", motmedelHttpErrors.ErrMultipleHeaderValues, name))
	}

	return headerValues[0], nil
}
