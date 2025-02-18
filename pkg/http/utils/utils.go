package utils

import (
	"bytes"
	"encoding/json"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

func DefaultRetryResponseChecker(response *http.Response, err error) bool {
	if response != nil {
		return response.StatusCode == 429 || response.StatusCode >= 500
	}

	if err != nil {
		return true
	}

	return false
}

type RetryConfiguration struct {
	Count           int
	BaseDelay       time.Duration
	MaximumWaitTime time.Duration
	CheckResponse   func(*http.Response, error) bool
}

type HttpRetryClient struct {
	http.Client
	RetryConfiguration *RetryConfiguration
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

func handleRequest(request *http.Request, httpClient HttpClient) (*http.Response, []byte, error) {
	if request == nil {
		return nil, nil, nil
	}

	if httpClient == nil {
		return nil, nil, motmedelHttpErrors.ErrNilHttpClient
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, nil, &motmedelErrors.Error{
			Message: "An error occurred when performing the request.",
			Cause:   err,
		}
	}
	if response == nil {
		return nil, nil, motmedelHttpErrors.ErrNilHttpResponse
	}
	responseBody := response.Body
	defer responseBody.Close()

	responseBodyData, err := io.ReadAll(responseBody)
	if err != nil {
		return response, nil, &motmedelErrors.Error{
			Message: "An error occurred when reading the response body.",
			Cause:   err,
		}
	}

	return response, responseBodyData, nil
}

func SendRequest(
	httpClient HttpClient,
	method string,
	url string,
	requestBody []byte,
	addToRequest func(*http.Request) error,
) (*motmedelHttpTypes.HttpContext, error) {
	if httpClient == nil {
		return nil, motmedelHttpErrors.ErrNilHttpClient
	}

	if method == "" {
		return nil, motmedelHttpErrors.ErrEmptyMethod
	}

	if url == "" {
		return nil, motmedelHttpErrors.ErrEmptyUrl
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, &motmedelErrors.Error{
			Message: "An error occurred when creating a request.",
			Cause:   err,
		}
	}
	if request == nil {
		return nil, motmedelHttpErrors.ErrNilHttpRequest
	}

	if addToRequest != nil {
		if err = addToRequest(request); err != nil {
			return nil, &motmedelErrors.Error{
				Message: "An error occurred when adding to a request.",
				Cause:   err,
			}
		}
	}

	var retryCount int
	var baseDelay time.Duration
	var maximumWaitTime time.Duration
	var checkRetryResponse func(*http.Response, error) bool

	httpRetryClient, ok := httpClient.(*HttpRetryClient)
	if ok && httpRetryClient != nil {
		if retryConfiguration := httpRetryClient.RetryConfiguration; retryConfiguration != nil {
			retryCount = max(retryCount, retryConfiguration.Count)
			baseDelay = retryConfiguration.BaseDelay
			maximumWaitTime = retryConfiguration.MaximumWaitTime
			checkRetryResponse = retryConfiguration.CheckResponse
		}
	}

	httpContext := &motmedelHttpTypes.HttpContext{Request: request, RequestBody: requestBody}
	var response *http.Response
	var responseBody []byte

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

		response, responseBody, err = handleRequest(request, httpClient)

		if checkRetryResponse == nil {
			checkRetryResponse = DefaultRetryResponseChecker
		}
		if !checkRetryResponse(response, err) {
			break
		}
		if err != nil && i != 0 {
			err = &motmedelHttpErrors.ReattemptFailedError{
				Cause:   err,
				Attempt: i + 1,
			}
		}
	}

	httpContext.Response = response

	if err != nil {
		return httpContext, &motmedelErrors.Error{
			Message: "An error occurred when handling an http request.",
			Cause:   err,
			Input:   request,
		}
	}
	httpContext.ResponseBody = responseBody

	if !strings.HasPrefix(strconv.Itoa(response.StatusCode), "2") {
		return httpContext, &motmedelHttpErrors.Non2xxStatusCodeError{StatusCode: response.StatusCode}
	}

	return httpContext, nil
}

func SendJsonRequestResponse[T any, U any](
	httpClient HttpClient,
	method string,
	url string,
	bodyValue *T,
	addToRequest func(*http.Request) error,
) (*U, *motmedelHttpTypes.HttpContext, error) {
	if httpClient == nil {
		return nil, nil, motmedelHttpErrors.ErrNilHttpClient
	}

	if method == "" {
		return nil, nil, motmedelHttpErrors.ErrEmptyMethod
	}

	if url == "" {
		return nil, nil, motmedelHttpErrors.ErrEmptyUrl
	}

	requestBody, err := json.Marshal(bodyValue)
	if err != nil {
		return nil, nil, &motmedelErrors.Error{
			Message: "An error occurred when marshalling the body value.",
			Cause:   err,
			Input:   bodyValue,
		}
	}

	httpContext, err := SendRequest(httpClient, method, url, requestBody, addToRequest)
	if err != nil {
		return nil, httpContext, &motmedelErrors.Error{
			Message: "An error occurred when sending the request.",
			Cause:   err,
			Input:   []any{httpClient, method, url, requestBody},
		}
	}
	if httpContext == nil {
		return nil, nil, nil
	}

	responseBody := httpContext.ResponseBody
	if len(responseBody) == 0 {
		return nil, httpContext, nil
	}

	var responseValue *U
	if err = json.Unmarshal(responseBody, &responseValue); err != nil {
		return nil, httpContext, &motmedelErrors.Error{
			Message: "An error occurred when unmarshalling the response body.",
			Cause:   err,
			Input:   responseBody,
		}
	}

	return responseValue, httpContext, nil
}
