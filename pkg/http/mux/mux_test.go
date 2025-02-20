package mux

import (
	"bytes"
	"encoding/json"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var httpServer *httptest.Server

func TestMain(m *testing.M) {
	mux := &Mux{}
	mux.Add(
		&endpoint_specification.EndpointSpecification{
			Path:   "/hello-world",
			Method: http.MethodGet,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return &muxTypesResponse.Response{Body: []byte("hello world")}, nil
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/empty",
			Method: http.MethodGet,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
		},
	)

	httpServer = httptest.NewServer(mux)

	code := m.Run()
	httpServer.Close()

	os.Exit(code)
}

func TestMux(t *testing.T) {
	testCases := []struct {
		name                  string
		url                   string
		method                string
		body                  []byte
		headers               [][2]string
		expectedStatusCode    int
		expectedHeaders       [][2]string
		expectedBody          []byte
		expectedProblemDetail *problem_detail.ProblemDetail
	}{
		{
			name:               "status ok",
			method:             http.MethodGet,
			url:                "/hello-world",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("hello world"),
		},
		{
			name:               "status no content",
			method:             http.MethodGet,
			url:                "/empty",
			expectedStatusCode: http.StatusNoContent,
		},
		// TODO: Test static content retrieval
		{
			name:               "error status not found",
			method:             http.MethodGet,
			url:                "/not-found",
			expectedStatusCode: http.StatusNotFound,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Type:   "about:blank",
				Title:  http.StatusText(http.StatusNotFound),
				Status: http.StatusNotFound,
			},
		},
		// TODO: Check if body is read (respond with request body)
		// TODO: Test method mismatch
		// TODO: Test body parser
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var requestBody io.Reader
			if testCaseBody := testCase.body; len(testCaseBody) != 0 {
				requestBody = bytes.NewReader(testCaseBody)
			}

			request, err := http.NewRequest(testCase.method, httpServer.URL+testCase.url, requestBody)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("http client do: %v", err)
			}
			defer response.Body.Close()

			responseBody, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("io read all response body: %v", err)
			}

			if response.StatusCode != testCase.expectedStatusCode {
				t.Errorf("got status code %d, expected %d", response.StatusCode, testCase.expectedStatusCode)
			}

			if expectedHeaders := testCase.expectedHeaders; len(expectedHeaders) != 0 {
				responseHeader := response.Header
				for _, header := range expectedHeaders {
					headerValue := responseHeader.Get(header[0])
					if headerValue != header[1] {
						t.Errorf("got %q, expected header %q to be %q", headerValue, header[0], header[1])
					}
				}
			}

			if expectedProblemDetail := testCase.expectedProblemDetail; expectedProblemDetail != nil {
				var problemDetail *problem_detail.ProblemDetail
				if err := json.Unmarshal(responseBody, &problemDetail); err != nil {
					t.Fatalf("json unmarshal response body: %v", err)
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(problem_detail.ProblemDetail{}, "Instance"),
					cmpopts.EquateEmpty(),
				}

				if diff := cmp.Diff(expectedProblemDetail, problemDetail, opts...); diff != "" {
					t.Errorf("problem detail mismatch (-expected +got):\n%s", diff)
				}
			} else if !bytes.Equal(responseBody, testCase.expectedBody) {
				t.Errorf("got response body %q, expected response body %q", responseBody, testCase.expectedBody)
			}
		})
	}
}
