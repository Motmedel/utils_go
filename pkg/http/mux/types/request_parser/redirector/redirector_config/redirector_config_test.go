package redirector_config

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
	if defaults.ParameterName != DefaultParameterName {
		t.Errorf("default parameter name = %q, want %q", defaults.ParameterName, DefaultParameterName)
	}
	if defaults.RequireProto != DefaultRequireProto {
		t.Errorf("default require proto = %v, want %v", defaults.RequireProto, DefaultRequireProto)
	}

	config := New(WithParameterName("next"), WithRequireProto(true))
	if config.ParameterName != "next" {
		t.Errorf("parameter name = %q, want %q", config.ParameterName, "next")
	}
	if !config.RequireProto {
		t.Error("expected RequireProto to be true")
	}
}

func TestNew_NilOptionIsIgnored(t *testing.T) {
	t.Parallel()

	if config := New(nil); config.ParameterName != DefaultParameterName {
		t.Errorf("nil option should be ignored, got %q", config.ParameterName)
	}
}
