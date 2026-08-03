package header_extractor_config

import (
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
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
		WithProblemDetailStatusCode(http.StatusUnauthorized),
		WithProblemDetailMissingText("missing"),
		WithProblemDetailMultipleText("multiple"),
	)
	if config.ProblemDetailStatusCode != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", config.ProblemDetailStatusCode, http.StatusUnauthorized)
	}
	if config.ProblemDetailMissingText != "missing" {
		t.Errorf("missing text = %q, want %q", config.ProblemDetailMissingText, "missing")
	}
	if config.ProblemDetailMultipleText != "multiple" {
		t.Errorf("multiple text = %q, want %q", config.ProblemDetailMultipleText, "multiple")
	}
}
