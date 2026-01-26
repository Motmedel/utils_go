package mux

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_loader"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_loader/body_setting"
	bodyParserAdapter "github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser/adapter"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser/json_body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint/static_content"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/firewall_verdict"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/rate_limiting"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/retry_after"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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
	detail *problem_detail.Detail,
	negotiation *motmedelHttpTypes.ContentNegotiation,
) ([]byte, string, error) {
	if detail == nil {
		return nil, "", nil_error.New("problem detail")
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
	mux.FirewallParser = request_parser.New(
		func(request *http.Request) (firewall_verdict.Verdict, *response_error.ResponseError) {
			if request.URL.RawQuery != "" {
				return firewall_verdict.Reject, &response_error.ResponseError{
					ProblemDetail: problem_detail.New(
						http.StatusForbidden,
						problem_detail_config.WithDetail("URL query parameters are not allowed."),
					),
				}
			}

			if request.URL.Path == "/secret-reject" {
				return firewall_verdict.Reject, nil
			}

			if request.URL.Path == "/secret-drop" {
				return firewall_verdict.Drop, nil
			}

			return firewall_verdict.Accept, nil
		},
	)
	mux.ProblemDetailConverter = HtmlProblemDetailConverter

	mux.Add(
		&endpoint.Endpoint{
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
		&endpoint.Endpoint{
			Path:   "/hello-world",
			Method: http.MethodPost,
		},
		&endpoint.Endpoint{
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
		&endpoint.Endpoint{
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
		&endpoint.Endpoint{
			Path:       "/push",
			Method:     http.MethodPost,
			BodyLoader: &body_loader.Loader{ContentType: "application/json"},
		},
		&endpoint.Endpoint{
			Path:   "/empty",
			Method: http.MethodGet,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
		},
		&endpoint.Endpoint{
			Path:   "/body-parsing",
			Method: http.MethodPost,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				d, responseError := utils.GetServerNonZeroParsedRequestBody[*bodyParserTestData](request.Context())
				if responseError != nil {
					return nil, responseError
				}

				if d.Data != "hello world" {
					return nil, &response_error.ResponseError{
						ProblemDetail: problem_detail.New(http.StatusInternalServerError),
					}
				}

				return nil, nil
			},
			BodyLoader: &body_loader.Loader{
				Parser:      bodyParserAdapter.New(json_body_parser.New[*bodyParserTestData]()),
				ContentType: "application/json",
			},
		},
		&endpoint.Endpoint{
			Path:   "/body-parsing-limit",
			Method: http.MethodPost,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
			BodyLoader: &body_loader.Loader{
				ContentType: "application/octet-stream",
				MaxBytes:    2,
			},
		},
		&endpoint.Endpoint{
			Path:   "/forbidden-body",
			Method: http.MethodPost,
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
			BodyLoader: &body_loader.Loader{
				Setting: body_setting.Forbidden,
			},
		},
		&endpoint.Endpoint{
			Path:   "/teapot",
			Method: http.MethodGet,
			Handler: func(request *http.Request, i []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, &response_error.ResponseError{
					ProblemDetail: problem_detail.New(http.StatusTeapot),
				}
			},
		},
		&endpoint.Endpoint{
			Path:   "/rate-limiting",
			Method: http.MethodGet,
			RateLimitingConfiguration: &rate_limiting.RateLimitingConfiguration{
				NumRequests:          3,
				NumSecondsExpiration: 5,
			},
		},
		&endpoint.Endpoint{
			Path:   "/cors",
			Method: http.MethodGet,
			CorsParser: request_parser.RequestParserFunction[*motmedelHttpTypes.CorsConfiguration](
				func(r *http.Request) (*motmedelHttpTypes.CorsConfiguration, *response_error.ResponseError) {
					return &motmedelHttpTypes.CorsConfiguration{
						Origin:        "*",
						Credentials:   true,
						Headers:       []string{"X-Custom-Header", "X-Custom-Header-2"},
						ExposeHeaders: []string{"X-Secret"},
					}, nil
				},
			),
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
		},
		&endpoint.Endpoint{
			Path:   "/cors",
			Method: http.MethodPost,
			CorsParser: request_parser.RequestParserFunction[*motmedelHttpTypes.CorsConfiguration](
				func(r *http.Request) (*motmedelHttpTypes.CorsConfiguration, *response_error.ResponseError) {
					return &motmedelHttpTypes.CorsConfiguration{
						Origin:        "*",
						Credentials:   true,
						Headers:       []string{"X-Custom-Header-3", "X-Custom-Header-4"},
						ExposeHeaders: []string{"X-Secret-2"},
					}, nil
				},
			),
			Handler: func(request *http.Request, body []byte) (*muxTypesResponse.Response, *response_error.ResponseError) {
				return nil, nil
			},
		},
		&endpoint.Endpoint{
			Path:   "/cors",
			Method: http.MethodPatch,
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
		expectedProblemDetail *problem_detail.Detail
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
			expectedProblemDetail: &problem_detail.Detail{
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
			expectedProblemDetail: &problem_detail.Detail{
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
			expectedProblemDetail: &problem_detail.Detail{},
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
			expectedProblemDetail: &problem_detail.Detail{
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
			expectedProblemDetail: &problem_detail.Detail{
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
			expectedProblemDetail: &problem_detail.Detail{
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
			expectedProblemDetail: &problem_detail.Detail{
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
			expectedProblemDetail: &problem_detail.Detail{Detail: "Invalid JSON body."},
		},
		{
			name:                  "error status forbidden, firewall match url query parameters",
			method:                http.MethodGet,
			url:                   "/foo?bar=fuu",
			expectedStatusCode:    http.StatusForbidden,
			expectedProblemDetail: &problem_detail.Detail{Detail: "URL query parameters are not allowed."},
		},
		{
			name:                  "error status forbidden, firewall match url secret (reject)",
			method:                http.MethodGet,
			url:                   "/secret-reject",
			expectedStatusCode:    http.StatusForbidden,
			expectedProblemDetail: &problem_detail.Detail{},
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
			expectedProblemDetail: &problem_detail.Detail{Detail: "Limit: 2 bytes"},
		},
		{
			name:               "max bytes ok",
			method:             http.MethodPost,
			url:                "/body-parsing-limit",
			headers:            [][2]string{{"Content-Type", "application/octet-stream"}},
			body:               []byte("12"),
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                  "forbidden body",
			method:                http.MethodPost,
			url:                   "/forbidden-body",
			body:                  []byte("12"),
			expectedStatusCode:    http.StatusRequestEntityTooLarge,
			expectedProblemDetail: &problem_detail.Detail{Detail: "Limit: 0 bytes"},
		},
		{
			name:                  "forbidden body get",
			method:                http.MethodGet,
			url:                   "/hello-world",
			body:                  []byte("12"),
			expectedStatusCode:    http.StatusRequestEntityTooLarge,
			expectedProblemDetail: &problem_detail.Detail{Detail: "Limit: 0 bytes"},
		},
		{
			name:   "cors preflight",
			method: http.MethodOptions,
			headers: [][2]string{
				{"Origin", "https://example.com"},
				{"Access-Control-Request-Method", "POST"},
				{"Access-Control-Request-Headers", "X-Custom-Header-3, X-Custom-Header-4"},
			},
			url:                "/cors",
			expectedStatusCode: http.StatusNoContent,
			expectedHeaders: [][2]string{
				{"Access-Control-Allow-Methods", "GET, HEAD, POST"},
				{"Access-Control-Allow-Origin", "*"},
				{"Access-Control-Allow-Credentials", "true"},
				{"Access-Control-Allow-Headers", "X-Custom-Header-3, X-Custom-Header-4"},
				{"Allow", "GET, HEAD, OPTIONS, PATCH, POST"},
			},
		},
		{
			name:   "cors post",
			method: http.MethodPost,
			headers: [][2]string{
				{"Origin", "https://example.com"},
			},
			url:                "/cors",
			expectedStatusCode: http.StatusNoContent,
			expectedHeaders: [][2]string{
				{"Access-Control-Allow-Origin", "*"},
				{"Access-Control-Allow-Credentials", "true"},
				{"Access-Control-Expose-Headers", "X-Secret-2"},
			},
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
				var problemDetail *problem_detail.Detail
				if err := json.Unmarshal(responseBody, &problemDetail); err != nil {
					t.Fatalf("json unmarshal response body: %v", err)
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(problem_detail.Detail{}, "Type"),
					cmpopts.IgnoreFields(problem_detail.Detail{}, "Instance"),
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
		retryAfter, err := retry_after.Parse([]byte(retryAfterValue))
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
