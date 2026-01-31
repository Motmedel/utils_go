package cookie_extractor

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/cookie_extractor/cookie_extractor_config"
)

func TestNew(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		_, err := New("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrEmptyName) {
			t.Errorf("got error %v, expected %v", err, ErrEmptyName)
		}
	})

	t.Run("valid name creates parser", func(t *testing.T) {
		parser, err := New("session")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}

		if parser.Name != "session" {
			t.Errorf("got name %q, expected %q", parser.Name, "session")
		}
	})

	t.Run("with custom options", func(t *testing.T) {
		parser, err := New(
			"auth-token",
			cookie_extractor_config.WithProblemDetailStatusCode(http.StatusUnauthorized),
			cookie_extractor_config.WithProblemDetailText("Missing authentication cookie."),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}
	})
}

func TestNewTokenCookieExtractor(t *testing.T) {
	t.Run("empty name returns error", func(t *testing.T) {
		_, err := NewTokenCookieExtractor("")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrEmptyName) {
			t.Errorf("got error %v, expected %v", err, ErrEmptyName)
		}
	})

	t.Run("creates parser with unauthorized status", func(t *testing.T) {
		parser, err := NewTokenCookieExtractor("token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if parser == nil {
			t.Fatal("expected non-nil parser")
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusUnauthorized
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}

		expectedDetail := "Missing token cookie."
		if responseError.ProblemDetail.Detail != expectedDetail {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedDetail)
		}
	})

	t.Run("options can override defaults", func(t *testing.T) {
		expectedDetail := "Custom token missing message."
		parser, err := NewTokenCookieExtractor(
			"token",
			cookie_extractor_config.WithProblemDetailText(expectedDetail),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		if responseError.ProblemDetail.Detail != expectedDetail {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedDetail)
		}
	})
}

func TestParser_Parse(t *testing.T) {
	t.Run("nil request returns error", func(t *testing.T) {
		parser, err := New("session")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		_, responseError := parser.Parse(nil)
		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequest) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, motmedelHttpErrors.ErrNilHttpRequest)
		}
	})

	t.Run("empty name in parser returns error", func(t *testing.T) {
		parser := &Parser{Name: ""}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ServerError == nil {
			t.Fatal("expected server error, got nil")
		}

		if !errors.Is(responseError.ServerError, ErrEmptyName) {
			t.Errorf("got error %v, expected %v", responseError.ServerError, ErrEmptyName)
		}
	})

	t.Run("missing cookie returns client error", func(t *testing.T) {
		parser, err := New("session")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ClientError == nil {
			t.Fatal("expected client error, got nil")
		}

		if !errors.Is(responseError.ClientError, http.ErrNoCookie) {
			t.Errorf("got error %v, expected %v", responseError.ClientError, http.ErrNoCookie)
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusBadRequest
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}

		extension := responseError.ProblemDetail.Extension
		if extension == nil {
			t.Fatal("expected problem detail extension, got nil")
		}

		cookieValue, ok := extension["cookie"]
		if !ok {
			t.Fatal("expected 'cookie' key in extension")
		}

		if cookieValue != "session" {
			t.Errorf("got cookie %q, expected %q", cookieValue, "session")
		}
	})

	t.Run("cookie present returns value", func(t *testing.T) {
		parser, err := New("session")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		expectedValue := "abc123"
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.AddCookie(&http.Cookie{Name: "session", Value: expectedValue})
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expectedValue {
			t.Errorf("got %q, expected %q", result, expectedValue)
		}
	})

	t.Run("correct cookie selected among multiple", func(t *testing.T) {
		parser, err := New("target")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		expectedValue := "target-value"
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.AddCookie(&http.Cookie{Name: "other1", Value: "other1-value"})
		request.AddCookie(&http.Cookie{Name: "target", Value: expectedValue})
		request.AddCookie(&http.Cookie{Name: "other2", Value: "other2-value"})
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expectedValue {
			t.Errorf("got %q, expected %q", result, expectedValue)
		}
	})

	t.Run("custom status code for missing cookie", func(t *testing.T) {
		parser, err := New(
			"auth-token",
			cookie_extractor_config.WithProblemDetailStatusCode(http.StatusUnauthorized),
		)
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		expectedStatusCode := http.StatusUnauthorized
		if responseError.ProblemDetail.Status != expectedStatusCode {
			t.Errorf("got status %d, expected %d", responseError.ProblemDetail.Status, expectedStatusCode)
		}
	})

	t.Run("custom text for missing cookie", func(t *testing.T) {
		expectedText := "Custom missing cookie message."
		parser, err := New(
			"session",
			cookie_extractor_config.WithProblemDetailText(expectedText),
		)
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, responseError := parser.Parse(request)

		if responseError == nil {
			t.Fatal("expected response error, got nil")
		}

		if responseError.ProblemDetail == nil {
			t.Fatal("expected problem detail, got nil")
		}

		if responseError.ProblemDetail.Detail != expectedText {
			t.Errorf("got detail %q, expected %q", responseError.ProblemDetail.Detail, expectedText)
		}
	})

	t.Run("empty cookie value is valid", func(t *testing.T) {
		parser, err := New("session")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.AddCookie(&http.Cookie{Name: "session", Value: ""})
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != "" {
			t.Errorf("got %q, expected empty string", result)
		}
	})

	t.Run("cookie with special characters", func(t *testing.T) {
		parser, err := New("session")
		if err != nil {
			t.Fatalf("unexpected error creating parser: %v", err)
		}

		expectedValue := "abc+123/def=ghi"
		request := httptest.NewRequest(http.MethodGet, "/test", nil)
		request.AddCookie(&http.Cookie{Name: "session", Value: expectedValue})
		result, responseError := parser.Parse(request)

		if responseError != nil {
			t.Fatalf("unexpected response error: %v", responseError)
		}

		if result != expectedValue {
			t.Errorf("got %q, expected %q", result, expectedValue)
		}
	})
}
