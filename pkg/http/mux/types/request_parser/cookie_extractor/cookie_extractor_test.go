package cookie_extractor

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse(t *testing.T) {
	t.Parallel()

	const cookieName = "session"

	newRequest := func(cookies ...*http.Cookie) *http.Request {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		for _, cookie := range cookies {
			request.AddCookie(cookie)
		}
		return request
	}

	parser, err := New(cookieName)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	tokenParser, err := NewTokenCookieExtractor("token")
	if err != nil {
		t.Fatalf("new token cookie extractor: %v", err)
	}

	testCases := []struct {
		name       string
		parser     *Parser
		request    *http.Request
		wantValue  string
		wantStatus int
		wantServer bool
	}{
		{
			name:      "present cookie",
			parser:    parser,
			request:   newRequest(&http.Cookie{Name: cookieName, Value: "abc", Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode}),
			wantValue: "abc",
		},
		{name: "missing cookie", parser: parser, request: newRequest(), wantStatus: http.StatusBadRequest},
		{name: "missing token cookie uses 401", parser: tokenParser, request: newRequest(), wantStatus: http.StatusUnauthorized},
		{name: "nil request", parser: parser, request: nil, wantServer: true},
		{name: "empty name", parser: &Parser{}, request: newRequest(), wantServer: true},
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
