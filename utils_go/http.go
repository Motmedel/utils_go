package utils_go

import (
	"bufio"
	"bytes"
	"net/http"
)

func ParseHttpRequestData(requestBytes []byte) (*http.Request, error) {
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(requestBytes)))
	if err != nil {
		// TODO: This could be an InputError? I could make some aggregated struct in `utils_go` + stack
		return nil, &CauseError{
			Message: "An error occurred when reading and parsing data as an HTTP request.",
			Cause:   err,
		}
	}
	return request, nil
}

func ParseHttpResponseData(responseBytes []byte) (*http.Response, error) {
	response, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(responseBytes)), nil)
	if err != nil {
		// TODO: This could be an InputError? I could make some aggregated struct in `utils_go` + stack
		return nil, &CauseError{
			Message: "An error occurred when reading and parsing data as an HTTP response.",
			Cause:   err,
		}
	}
	return response, nil
}
