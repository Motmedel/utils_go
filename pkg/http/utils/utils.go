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
)

func SendRequest(
	httpClient *http.Client,
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
		return nil, &motmedelErrors.CauseError{
			Message: "An error occurred when creating a request.",
			Cause:   err,
		}
	}
	if request == nil {
		return nil, motmedelHttpErrors.ErrNilHttpRequest
	}

	if addToRequest != nil {
		if err = addToRequest(request); err != nil {
			return nil, &motmedelErrors.CauseError{
				Message: "An error occurred when adding to a request.",
				Cause:   err,
			}
		}
	}

	httpContext := &motmedelHttpTypes.HttpContext{Request: request, RequestBody: requestBody}

	response, err := httpClient.Do(request)
	if err != nil {
		return httpContext, &motmedelErrors.CauseError{
			Message: "An error occurred when performing the request.",
			Cause:   err,
		}
	}
	if response == nil {
		return httpContext, motmedelHttpErrors.ErrNilHttpResponse
	}
	defer response.Body.Close()
	httpContext.Response = response

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return httpContext, &motmedelErrors.CauseError{
			Message: "An error occurred when reading the response body.",
			Cause:   err,
		}
	}
	httpContext.ResponseBody = responseBody

	if !strings.HasPrefix(strconv.Itoa(response.StatusCode), "2") {
		return httpContext, &motmedelHttpErrors.Non2xxStatusCodeError{StatusCode: response.StatusCode}
	}

	return httpContext, nil
}

func SendJsonRequestResponse[T any, U any](
	httpClient *http.Client,
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
		return nil, nil, &motmedelErrors.InputError{
			Message: "An error occurred when marshalling the body value.",
			Cause:   err,
			Input:   bodyValue,
		}
	}

	httpContext, err := SendRequest(httpClient, method, url, requestBody, addToRequest)
	if err != nil {
		return nil, httpContext, &motmedelErrors.InputError{
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
		return nil, httpContext, &motmedelErrors.InputError{
			Message: "An error occurred when unmarshalling the response body.",
			Cause:   err,
			Input:   responseBody,
		}
	}

	return responseValue, httpContext, nil
}
