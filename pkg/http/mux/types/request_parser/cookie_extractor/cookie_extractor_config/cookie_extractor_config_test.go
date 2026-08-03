package cookie_extractor_config

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
	if defaults.ProblemDetailText != DefaultProblemDetailText {
		t.Errorf("default text = %q, want %q", defaults.ProblemDetailText, DefaultProblemDetailText)
	}

	config := New(
		WithProblemDetailStatusCode(http.StatusForbidden),
		WithProblemDetailText("nope"),
	)
	if config.ProblemDetailStatusCode != http.StatusForbidden {
		t.Errorf("status code = %d, want %d", config.ProblemDetailStatusCode, http.StatusForbidden)
	}
	if config.ProblemDetailText != "nope" {
		t.Errorf("text = %q, want %q", config.ProblemDetailText, "nope")
	}
}
