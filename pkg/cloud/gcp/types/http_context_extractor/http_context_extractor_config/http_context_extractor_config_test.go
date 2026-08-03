package http_context_extractor_config

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
	if defaults.ProjectId != "" {
		t.Errorf("default ProjectId = %q, want empty", defaults.ProjectId)
	}

	config := New(WithProjectId("my-project"))
	if config.ProjectId != "my-project" {
		t.Errorf("ProjectId = %q, want %q", config.ProjectId, "my-project")
	}
}

func TestNewSkipsNilOption(t *testing.T) {
	t.Parallel()

	// New must ignore nil options rather than panic.
	config := New(nil, WithProjectId("p"), nil)
	if config.ProjectId != "p" {
		t.Errorf("ProjectId = %q, want %q", config.ProjectId, "p")
	}
}
