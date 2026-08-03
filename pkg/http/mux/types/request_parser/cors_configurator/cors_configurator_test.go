package cors_configurator

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRequestWithOrigin(t *testing.T, origin string) *http.Request {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	return request
}

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("nil request is a server error", func(t *testing.T) {
		t.Parallel()
		_, responseError := (&Configurator{}).Parse(nil)
		if responseError == nil || responseError.ServerError == nil {
			t.Fatalf("expected a server error, got %#v", responseError)
		}
	})

	t.Run("no origin yields no config", func(t *testing.T) {
		t.Parallel()
		configurator := &Configurator{AllowedOrigins: []string{"https://example.com"}}
		config, responseError := configurator.Parse(newRequestWithOrigin(t, ""))
		if responseError != nil {
			t.Fatalf("unexpected error: %#v", responseError)
		}
		if config != nil {
			t.Fatalf("expected nil config, got %#v", config)
		}
	})

	t.Run("allowed origin matches case-insensitively", func(t *testing.T) {
		t.Parallel()
		configurator := &Configurator{AllowedOrigins: []string{"https://example.com"}, Credentials: true}
		config, responseError := configurator.Parse(newRequestWithOrigin(t, "https://EXAMPLE.com"))
		if responseError != nil {
			t.Fatalf("unexpected error: %#v", responseError)
		}
		if config == nil || config.Origin != "https://example.com" {
			t.Fatalf("expected the allowed origin, got %#v", config)
		}
		if !config.Credentials {
			t.Error("expected Credentials to be carried over")
		}
	})

	t.Run("disallowed origin yields no config", func(t *testing.T) {
		t.Parallel()
		configurator := &Configurator{AllowedOrigins: []string{"https://example.com"}}
		config, responseError := configurator.Parse(newRequestWithOrigin(t, "https://evil.com"))
		if responseError != nil {
			t.Fatalf("unexpected error: %#v", responseError)
		}
		if config != nil {
			t.Fatalf("expected nil config, got %#v", config)
		}
	})

	t.Run("registered domain match", func(t *testing.T) {
		t.Parallel()
		configurator := &Configurator{RegisteredDomain: "example.com"}
		config, responseError := configurator.Parse(newRequestWithOrigin(t, "https://app.example.com"))
		if responseError != nil {
			t.Fatalf("unexpected error: %#v", responseError)
		}
		if config == nil || config.Origin != "https://app.example.com" {
			t.Fatalf("expected a registered-domain match, got %#v", config)
		}
	})
}
