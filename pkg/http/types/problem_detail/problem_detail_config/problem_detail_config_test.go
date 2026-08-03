package problem_detail_config

import (
	"testing"
)

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	config := New()

	if config.Type != "" {
		t.Errorf("Type = %q, want empty", config.Type)
	}
	if config.Detail != "" {
		t.Errorf("Detail = %q, want empty", config.Detail)
	}
	if config.Extension != nil {
		t.Errorf("Extension = %v, want nil", config.Extension)
	}
	if len(config.Instance) != 36 {
		t.Errorf("Instance = %q (len %d), want a 36-char UUID string", config.Instance, len(config.Instance))
	}
}

func TestNewInstanceUnique(t *testing.T) {
	t.Parallel()

	first := New()
	second := New()
	if first.Instance == second.Instance {
		t.Errorf("two New() calls produced identical Instance %q", first.Instance)
	}
}

func TestNewOptions(t *testing.T) {
	t.Parallel()

	extension := map[string]any{"key": "value"}

	config := New(
		WithType("https://example.com/type"),
		WithInstance("fixed-instance"),
		WithDetail("some detail"),
		WithExtension(extension),
	)

	if config.Type != "https://example.com/type" {
		t.Errorf("Type = %q, want %q", config.Type, "https://example.com/type")
	}
	if config.Instance != "fixed-instance" {
		t.Errorf("Instance = %q, want %q", config.Instance, "fixed-instance")
	}
	if config.Detail != "some detail" {
		t.Errorf("Detail = %q, want %q", config.Detail, "some detail")
	}
	if config.Extension["key"] != "value" {
		t.Errorf("Extension = %v, want %v", config.Extension, extension)
	}
}
