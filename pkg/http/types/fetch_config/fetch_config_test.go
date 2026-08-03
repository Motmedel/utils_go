package fetch_config

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config/retry_config"
)

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	config := New()

	if config.Method != DefaultMethod {
		t.Errorf("Method = %q, want %q", config.Method, DefaultMethod)
	}
	if config.HttpClient != http.DefaultClient {
		t.Errorf("HttpClient = %p, want %p", config.HttpClient, http.DefaultClient)
	}
	if config.Headers != nil {
		t.Errorf("Headers = %v, want nil", config.Headers)
	}
	if config.Body != nil {
		t.Errorf("Body = %v, want nil", config.Body)
	}
	if config.SkipReadResponseBody {
		t.Error("SkipReadResponseBody = true, want false")
	}
	if config.SkipErrorOnStatus {
		t.Error("SkipErrorOnStatus = true, want false")
	}
	if config.RetryConfig != nil {
		t.Errorf("RetryConfig = %v, want nil", config.RetryConfig)
	}
}

func TestNewNilOptionSkipped(t *testing.T) {
	t.Parallel()

	config := New(nil, WithMethod(http.MethodPost), nil)
	if config.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", config.Method, http.MethodPost)
	}
}

func TestNewOptions(t *testing.T) {
	t.Parallel()

	headers := map[string]string{"X-Test": "value"}
	body := []byte("payload")
	retryConfig := retry_config.New()
	httpClient := &http.Client{}

	config := New(
		WithMethod(http.MethodPost),
		WithHeaders(headers),
		WithBody(body),
		WithSkipReadResponseBody(true),
		WithSkipErrorOnStatus(true),
		WithRetryConfig(retryConfig),
		WithHttpClient(httpClient),
	)

	if config.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", config.Method, http.MethodPost)
	}
	if config.Headers["X-Test"] != "value" {
		t.Errorf("Headers = %v, want %v", config.Headers, headers)
	}
	if !bytes.Equal(config.Body, body) {
		t.Errorf("Body = %q, want %q", config.Body, body)
	}
	if !config.SkipReadResponseBody {
		t.Error("SkipReadResponseBody = false, want true")
	}
	if !config.SkipErrorOnStatus {
		t.Error("SkipErrorOnStatus = false, want true")
	}
	if config.RetryConfig != retryConfig {
		t.Errorf("RetryConfig = %p, want %p", config.RetryConfig, retryConfig)
	}
	if config.HttpClient != httpClient {
		t.Errorf("HttpClient = %p, want %p", config.HttpClient, httpClient)
	}
}
