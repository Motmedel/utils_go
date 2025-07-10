package mux

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/firewall"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/rate_limiting"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	bodyParserJson "github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser/json"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/retry_after"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	problemDetailErrors "github.com/Motmedel/utils_go/pkg/http/problem_detail/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"
	"time"
)

var httpServer *httptest.Server

type bodyParserTestData struct {
	Data string `json:"data"`
}

var defaultHtmlProblemDetailMediaRanges = slices.Concat(
	response_error.DefaultProblemDetailMediaRanges,
	[]*motmedelHttpTypes.ServerMediaRange{{Type: "text", Subtype: "html"}},
)

func htmlConvertProblemDetail(
	detail *problem_detail.ProblemDetail,
	negotiation *motmedelHttpTypes.ContentNegotiation,
) ([]byte, string, error) {
	if detail == nil {
		return nil, "", motmedelErrors.NewWithTrace(problemDetailErrors.ErrNilProblemDetail)
	}

	if detail.Status == http.StatusTeapot && negotiation != nil {
		if negotiation.NegotiatedAccept == "" {
			matchingServerMediaRange := motmedelHttpUtils.GetMatchingAccept(
				negotiation.Accept.GetPriorityOrderedEncodings(),
				defaultHtmlProblemDetailMediaRanges,
			)
			if matchingServerMediaRange != nil {
				negotiation.NegotiatedAccept = matchingServerMediaRange.GetFullType(true)
			}
		}

		switch negotiatedAccept := negotiation.NegotiatedAccept; negotiatedAccept {
		case "text/html":
			return []byte(fmt.Sprintf("<html>%d</html>", detail.Status)), "text/html", nil
		}
	}

	return response_error.ConvertProblemDetail(detail, negotiation)
}

var HtmlProblemDetailConverter = response_error.ProblemDetailConverterFunction(htmlConvertProblemDetail)

func TestMain(m *testing.M) {
	mux := &Mux{}
	mux.FirewallConfiguration = &firewall.Configuration{
		Handler: func(request *http.Request) (firewall.Verdict, *response_error.ResponseError) {
			if request.URL.RawQuery != "" {
				return firewall.VerdictReject, &response_error.ResponseError{
					ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
						http.StatusForbidden,
						"URL query parameters are not allowed.",
						nil,
					),
				}
			}

			if request.URL.Path == "/secret-reject" {
				return firewall.VerdictReject, nil
			}

			if request.URL.Path == "/secret-drop" {
				return firewall.VerdictDrop, nil
			}

			return firewall.VerdictAccept, nil
		},
	}
	mux.ProblemDetailConverter = HtmlProblemDetailConverter

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
			Path:   "/hello-world",
			Method: http.MethodPost,
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
			Path:   "/hello-world-fetch-metadata",
			Method: http.MethodGet,
			StaticContent: &muxTypesStaticContent.StaticContent{
				StaticContentData: muxTypesStaticContent.StaticContentData{
					Data: []byte("<html>hello world</html>"),
					Headers: []*muxTypesResponse.HeaderEntry{
						{Name: "Content-Type", Value: "text/html"},
						{Name: "Cache-Control", Value: "no-cache", Overwrite: true},
					},
				},
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/push",
			Method: http.MethodPost,
			BodyParserConfiguration: &parsing.BodyParserConfiguration{
				ContentType: "application/json",
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/empty",
			Method: http.MethodGet,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/body-parsing",
			Method: http.MethodPost,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				d, ok := request.Context().Value(parsing.ParsedRequestBodyContextKey).(*bodyParserTestData)
				if !ok || d == nil {
					return nil, &response_error.ResponseError{
						ServerError: motmedelErrors.NewWithTrace(
							errors.New("could not obtain parsed request body"),
						),
					}
				}

				if d.Data != "hello world" {
					return nil, &response_error.ResponseError{
						ProblemDetail: problem_detail.MakeInternalServerErrorProblemDetail("", nil),
					}
				}

				return nil, nil
			},
			BodyParserConfiguration: &parsing.BodyParserConfiguration{
				ContentType: "application/json",
				Parser:      bodyParserJson.New[bodyParserTestData](),
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/body-parsing-limit",
			Method: http.MethodPost,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
			BodyParserConfiguration: &parsing.BodyParserConfiguration{
				ContentType: "application/octet-stream",
				MaxBytes:    2,
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/teapot",
			Method: http.MethodGet,
			Handler: func(request *http.Request, i []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, &response_error.ResponseError{
					ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(http.StatusTeapot, "", nil),
				}
			},
		},
		&endpoint_specification.EndpointSpecification{
			Path:   "/rate-limiting",
			Method: http.MethodGet,
			RateLimitingConfiguration: &rate_limiting.RateLimitingConfiguration{
				NumRequests:          3,
				NumSecondsExpiration: 5,
			},
		},
	)

	httpServer = httptest.NewServer(mux)

	code := m.Run()
	httpServer.Close()

	os.Exit(code)
}

// TODO: Test adding and deleting from a mux.

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
		expectedClientDoError error
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
			name:               "status no content (OPTIONS)",
			method:             http.MethodOptions,
			url:                "/hello-world",
			expectedStatusCode: http.StatusNoContent,
			expectedHeaders:    [][2]string{{"Allow", "GET, HEAD, OPTIONS, POST"}},
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
			name:               "fetch metadata vary",
			method:             http.MethodHead,
			url:                "/hello-world-fetch-metadata",
			expectedStatusCode: http.StatusOK,
			expectedHeaders:    [][2]string{{"Vary", "Sec-Fetch-Dest, Sec-Fetch-Mode, Sec-Fetch-Site"}},
		},
		{
			name:   "fetch metadata forbidden",
			method: http.MethodGet,
			url:    "/hello-world",
			headers: [][2]string{
				{"Sec-Fetch-Site", "cross-origin"},
			},
			expectedStatusCode: http.StatusForbidden,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: "Cross-site request blocked by Fetch-Metadata policy.",
			},
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
			method:             http.MethodPatch,
			url:                "/hello-world",
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedProblemDetail: &problem_detail.ProblemDetail{
				Detail: `Expected GET, HEAD, OPTIONS, POST.`,
			},
			expectedHeaders: [][2]string{{"Allow", "GET, HEAD, OPTIONS, POST"}},
		},
		{
			name:               "error status method not allowed, without response body",
			method:             http.MethodPatch,
			url:                "/hello-world",
			headers:            [][2]string{{"Accept-Encoding", "*;q=0"}},
			expectedStatusCode: http.StatusMethodNotAllowed,
			expectedHeaders:    [][2]string{{"Allow", "GET, HEAD, OPTIONS, POST"}},
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
			name:               "status no content, body parsing test",
			method:             http.MethodPost,
			url:                "/body-parsing",
			headers:            [][2]string{{"Content-Type", "application/json"}},
			body:               []byte(`{"data": "hello world"}`),
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
			name:                  "error bad request, invalid json body",
			method:                http.MethodPost,
			url:                   "/push",
			headers:               [][2]string{{"Content-Type", "application/json"}},
			body:                  []byte(`{"data": "data"`),
			expectedStatusCode:    http.StatusBadRequest,
			expectedProblemDetail: &problem_detail.ProblemDetail{Detail: "Invalid JSON body."},
		},
		{
			name:                  "error status forbidden, firewall match url query parameters",
			method:                http.MethodGet,
			url:                   "/foo?bar=fuu",
			expectedStatusCode:    http.StatusForbidden,
			expectedProblemDetail: &problem_detail.ProblemDetail{Detail: "URL query parameters are not allowed."},
		},
		{
			name:                  "error status forbidden, firewall match url secret (reject)",
			method:                http.MethodGet,
			url:                   "/secret-reject",
			expectedStatusCode:    http.StatusForbidden,
			expectedProblemDetail: &problem_detail.ProblemDetail{},
		},
		{
			name:                  "error status forbidden, firewall match url secret (drop)",
			method:                http.MethodGet,
			url:                   "/secret-drop",
			expectedClientDoError: io.EOF,
		},
		{
			name:               "custom error representation",
			method:             http.MethodGet,
			url:                "/teapot",
			headers:            [][2]string{{"Accept", "text/html"}},
			expectedStatusCode: http.StatusTeapot,
			expectedBody:       []byte("<html>418</html>"),
			expectedHeaders:    [][2]string{{"Content-Type", "text/html"}},
		},
		{
			name:                  "max bytes too large",
			method:                http.MethodPost,
			url:                   "/body-parsing-limit",
			headers:               [][2]string{{"Content-Type", "application/octet-stream"}},
			body:                  []byte("123"),
			expectedStatusCode:    http.StatusRequestEntityTooLarge,
			expectedProblemDetail: &problem_detail.ProblemDetail{Detail: "Limit: 2 bytes"},
		},
		{
			name:               "max bytes ok",
			method:             http.MethodPost,
			url:                "/body-parsing-limit",
			headers:            [][2]string{{"Content-Type", "application/octet-stream"}},
			body:               []byte("12"),
			expectedStatusCode: http.StatusNoContent,
		},
	}

	// TODO: Write test for URL parsing.

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
				if testCase.expectedClientDoError != nil {
					if errors.Is(err, testCase.expectedClientDoError) {
						return
					}
					t.Fatalf("http client do: %v", err)
				}
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

func TestRateLimiting(t *testing.T) {
	path := "/rate-limiting"

	for i := 0; i < 3; i++ {
		response, err := http.Get(httpServer.URL + path)
		if err != nil {
			t.Fatalf("http get: %v", err)
		}

		if response.StatusCode != http.StatusNoContent {
			t.Errorf("got status code %d, expected %d", response.StatusCode, http.StatusNoContent)
		}
	}

	response, err := http.Get(httpServer.URL + path)
	if err != nil {
		t.Fatalf("http get: %v", err)
	}

	expectedStatusCode := http.StatusTooManyRequests
	if response.StatusCode != expectedStatusCode {
		t.Errorf("got status code %d, expected %d", response.StatusCode, expectedStatusCode)
	}

	retryAfterValue := response.Header.Get("Retry-After")
	if retryAfterValue == "" {
		t.Error("no Retry-After header")
	} else {
		retryAfter, err := retry_after.ParseRetryAfter([]byte(retryAfterValue))
		if err != nil {
			t.Errorf("invalid Retry-After: %v", err)
		} else {
			waitTime, ok := retryAfter.WaitTime.(time.Time)
			if !ok {
				t.Error("invalid Retry-After wait time")
			}

			time.Sleep(time.Until(waitTime))

			response, err = http.Get(httpServer.URL + path)
			if err != nil {
				t.Fatalf("http get: %v", err)
			}

			if response.StatusCode != http.StatusNoContent {
				t.Errorf("got status code %d, expected %d", response.StatusCode, http.StatusNoContent)
			}
		}
	}
}
