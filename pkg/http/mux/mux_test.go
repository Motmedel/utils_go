package mux

import (
	"bytes"
	"encoding/json"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
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
				return &muxTypesResponse.Response{
					Body: []byte("hello world"),
					Headers: []*muxTypesResponse.HeaderEntry{
						{Name: "Content-Type", Value: "application/octet-stream"},
					},
				}, nil
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/hello-world-static",
			Method: http.MethodGet,
			StaticContent: &muxTypesStaticContent.StaticContent{
				StaticContentData: muxTypesStaticContent.StaticContentData{
					Data: []byte("<html>hello world</html>"),
					Headers: []*muxTypesResponse.HeaderEntry{
						{Name: "Content-Type", Value: "text/html"},
					},
				},
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/push",
			Method: http.MethodPost,
			BodyParserConfiguration: &parsing.BodyParserConfiguration{
				ContentType: "application/json",
				AllowEmpty:  false,
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
		method                string
		url                   string
		headers               [][2]string
		body                  []byte
		expectedStatusCode    int
		expectedHeaders       [][2]string
		expectedBody          []byte
		expectedProblemDetail *problem_detail.ProblemDetail
	}{
		{
			name:               "status ok, handler",
			method:             http.MethodGet,
			url:                "/hello-world",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("hello world"),
			expectedHeaders:    [][2]string{{"Content-Type", "application/octet-stream"}},
		},
		{
			name:               "status ok, handler (HEAD)",
			method:             http.MethodHead,
			url:                "/hello-world",
			expectedStatusCode: http.StatusOK,
			expectedHeaders:    [][2]string{{"Content-Type", "application/octet-stream"}},
		},
		{
			name:               "status ok, static content",
			method:             http.MethodGet,
			url:                "/hello-world-static",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("<html>hello world</html>"),
			expectedHeaders:    [][2]string{{"Content-Type", "text/html"}},
		},
		{
			name:               "default headers",
			method:             http.MethodGet,
			url:                "/hello-world",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("hello world"),
			expectedHeaders: func() [][2]string {
				var headers [][2]string
				for key, value := range response_writer.DefaultHeaders {
					headers = append(headers, [2]string{key, value})
				}
				return headers
			}(),
		},
		{
			name:               "default document headers",
			method:             http.MethodGet,
			url:                "/hello-world-static",
			expectedStatusCode: http.StatusOK,
			expectedBody:       []byte("<html>hello world</html>"),
			expectedHeaders: func() [][2]string {
				var headers [][2]string
				for key, value := range response_writer.DefaultHeaders {
					headers = append(headers, [2]string{key, value})
				}

				for key, value := range response_writer.DefaultDocumentHeaders {
					headers = append(headers, [2]string{key, value})
				}
				return headers
			}(),
		},
		{
			name:               "bad method",
			method:             http.MethodPost,
			url:                "/hello-world",
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: `Expected GET, HEAD, OPTIONS; observed "POST".`,
			},
			expectedHeaders: [][2]string{{"Allow", "GET, HEAD, OPTIONS"}},
		},
		{
			name:               "error status method not allowed, without response body",
			method:             http.MethodPost,
			url:                "/hello-world",
			headers:            [][2]string{{"Accept-Encoding", "*;q=0"}},
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedHeaders:    [][2]string{{"Allow", "GET, HEAD, OPTIONS"}},
		},
		{
			name:               "status no content",
			method:             http.MethodGet,
			url:                "/empty",
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                  "error status not found",
			method:                http.MethodGet,
			url:                   "/not-found",
			expectedStatusCode:    http.StatusNotFound,
			expectedProblemDetail: &problem_detail.ProblemDetail{},
		},
		{
			name:               "status ok, post data",
			method:             http.MethodPost,
			url:                "/push",
			headers:            [][2]string{{"Content-Type", "application/json"}},
			body:               []byte(`{"data": "data"}`),
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "error status bad request, post no data",
			method:             http.MethodPost,
			url:                "/push",
			headers:            [][2]string{{"Content-Type", "application/json"}},
			expectedStatusCode: http.StatusBadRequest,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: "A body is expected; Content-Length cannot be 0.",
			},
		},
		{
			name:               "error status unsupported media type, missing content type",
			method:             http.MethodPost,
			url:                "/push",
			body:               []byte(`{"data": "data"}`),
			expectedStatusCode: http.StatusUnsupportedMediaType,
			expectedHeaders:    [][2]string{{"Accept", "application/json"}},
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: "Missing Content-Type.",
			},
		},
		{
			name:               "error status bad request, malformed content type",
			method:             http.MethodPost,
			url:                "/push",
			headers:            [][2]string{{"Content-Type", ""}},
			body:               []byte(`{"data": "data"}`),
			expectedStatusCode: http.StatusBadRequest,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: "Malformed Content-Type.",
			},
		},
		{
			name:               "error status unsupported media type, other content type",
			method:             http.MethodPost,
			url:                "/push",
			headers:            [][2]string{{"Content-Type", "text/plain"}},
			expectedStatusCode: http.StatusUnsupportedMediaType,
			expectedHeaders:    [][2]string{{"Accept", "application/json"}},
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: `Expected Content-Type to be "application/json", observed "text/plain".`,
			},
		},
		{
			name:               "error bad request, invalid json body",
			method:             http.MethodPost,
			url:                "/push",
			headers:            [][2]string{{"Content-Type", "application/json"}},
			body:               []byte(`{"data": "data"`),
			expectedStatusCode: http.StatusBadRequest,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: `Invalid JSON body.`,
			},
		},
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

			for _, header := range testCase.headers {
				request.Header.Set(header[0], header[1])
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
					cmpopts.IgnoreFields(problem_detail.ProblemDetail{}, "Type"),
					cmpopts.IgnoreFields(problem_detail.ProblemDetail{}, "Instance"),
					cmpopts.EquateEmpty(),
				}

				expectedStatusCode := testCase.expectedStatusCode
				expectedProblemDetail.Title = http.StatusText(expectedStatusCode)
				expectedProblemDetail.Status = expectedStatusCode

				if diff := cmp.Diff(expectedProblemDetail, problemDetail, opts...); diff != "" {
					t.Errorf("problem detail mismatch (-expected +got):\n%s", diff)
				}
			} else if !bytes.Equal(responseBody, testCase.expectedBody) {
				t.Errorf("got response body %q, expected response body %q", responseBody, testCase.expectedBody)
			}
		})
	}
}
