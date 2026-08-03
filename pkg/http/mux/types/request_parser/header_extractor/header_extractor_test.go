package header_extractor

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	const headerName = "X-Token"

	newRequest := func(values ...string) *http.Request {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		request.Header = http.Header{}
		for _, value := range values {
			request.Header.Add(headerName, value)
		}
		return request
	}

	parser, err := New(headerName)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	testCases := []struct {
		name       string
		parser     *Parser
		request    *http.Request
		wantValue  string
		wantStatus int
		wantServer bool
	}{
		{name: "single value", parser: parser, request: newRequest("secret"), wantValue: "secret"},
		{name: "missing header", parser: parser, request: newRequest(), wantStatus: http.StatusBadRequest},
		{name: "multiple values", parser: parser, request: newRequest("a", "b"), wantStatus: http.StatusBadRequest},
		{name: "nil request", parser: parser, request: nil, wantServer: true},
		{name: "nil header", parser: parser, request: &http.Request{}, wantServer: true},
		{name: "empty name", parser: &Parser{}, request: newRequest("x"), wantServer: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			value, responseError := testCase.parser.Parse(testCase.request)

			switch {
			case testCase.wantServer:
				if responseError == nil || responseError.ServerError == nil {
					t.Fatalf("expected a server error, got %#v", responseError)
				}
			case testCase.wantStatus != 0:
				if responseError == nil || responseError.ProblemDetail == nil {
					t.Fatalf("expected a problem detail, got %#v", responseError)
				}
				if responseError.ProblemDetail.Status != testCase.wantStatus {
					t.Fatalf("expected status %d, got %d", testCase.wantStatus, responseError.ProblemDetail.Status)
				}
			default:
				if responseError != nil {
					t.Fatalf("unexpected error: %#v", responseError)
				}
				if value != testCase.wantValue {
					t.Fatalf("got %q, want %q", value, testCase.wantValue)
				}
			}
		})
	}
}
