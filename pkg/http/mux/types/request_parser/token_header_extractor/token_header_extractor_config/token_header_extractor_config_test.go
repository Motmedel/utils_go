package token_header_extractor_config

import (
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
	if defaults.HeaderName != DefaultHeaderName {
		t.Errorf("default header name = %q, want %q", defaults.HeaderName, DefaultHeaderName)
	}
	if defaults.HeaderValuePrefix != DefaultHeaderValuePrefix {
		t.Errorf("default header value prefix = %q, want %q", defaults.HeaderValuePrefix, DefaultHeaderValuePrefix)
	}
	if defaults.ProblemDetailStatusCode != DefaultProblemDetailStatusCode {
		t.Errorf("default status code = %d, want %d", defaults.ProblemDetailStatusCode, DefaultProblemDetailStatusCode)
	}
	if defaults.ProblemDetailMissingText != DefaultProblemDetailMissingText {
		t.Errorf("default missing text = %q, want %q", defaults.ProblemDetailMissingText, DefaultProblemDetailMissingText)
	}
	if defaults.ProblemDetailMultipleText != DefaultProblemDetailMultipleText {
		t.Errorf("default multiple text = %q, want %q", defaults.ProblemDetailMultipleText, DefaultProblemDetailMultipleText)
	}

	config := New(
		WithHeaderName("X-Token"),
		WithHeaderValuePrefix("Token "),
		WithProblemDetailStatusCode(http.StatusForbidden),
		WithProblemDetailMissingText("missing"),
		WithProblemDetailMultipleText("multiple"),
	)
	if config.HeaderName != "X-Token" {
		t.Errorf("header name = %q, want %q", config.HeaderName, "X-Token")
	}
	if config.HeaderValuePrefix != "Token " {
		t.Errorf("header value prefix = %q, want %q", config.HeaderValuePrefix, "Token ")
	}
	if config.ProblemDetailStatusCode != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", config.ProblemDetailStatusCode, http.StatusForbidden)
	}
	if config.ProblemDetailMissingText != "missing" {
		t.Errorf("missing text = %q, want %q", config.ProblemDetailMissingText, "missing")
	}
	if config.ProblemDetailMultipleText != "multiple" {
		t.Errorf("multiple text = %q, want %q", config.ProblemDetailMultipleText, "multiple")
	}
}
