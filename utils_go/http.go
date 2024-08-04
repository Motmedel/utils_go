package utils_go

import (
	"bufio"
	"bytes"
	"net/http"
)

func ParseHttpRequestData(requestBytes []byte) (*http.Request, error) {
	if len(requestBytes) == 0 {
		return nil, nil
	}

	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(requestBytes)))
	if err != nil {
		return nil, &InputError{
			Message: "An error occurred when reading and parsing data as an HTTP request.",
			Cause:   err,
			Input:   requestBytes,
		}
	}
	return request, nil
}

func ParseHttpResponseData(responseBytes []byte) (*http.Response, error) {
	if len(responseBytes) == 0 {
		return nil, nil
	}

	response, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(responseBytes)), nil)
	if err != nil {
		return nil, &InputError{
			Message: "An error occurred when reading and parsing data as an HTTP response.",
			Cause:   err,
			Input:   responseBytes,
		}
	}
	return response, nil
}
