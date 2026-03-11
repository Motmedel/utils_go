package logger

import (
	"log/slog"
	"testing"
)

func TestReplaceAttr_TimeKey(t *testing.T) {
	attr := ReplaceAttr(nil, slog.String(slog.TimeKey, "2024-01-01"))
	if attr.Key != "time" {
		t.Errorf("expected key 'time', got %q", attr.Key)
	}
}

func TestReplaceAttr_LevelKey(t *testing.T) {
	attr := ReplaceAttr(nil, slog.String(slog.LevelKey, "INFO"))
	if attr.Key != "severity" {
		t.Errorf("expected key 'severity', got %q", attr.Key)
	}
}

func TestReplaceAttr_MessageKey(t *testing.T) {
	attr := ReplaceAttr(nil, slog.String(slog.MessageKey, "hello"))
	if attr.Key != "message" {
		t.Errorf("expected key 'message', got %q", attr.Key)
	}
}

func TestReplaceAttr_SourceKey(t *testing.T) {
	source := &slog.Source{File: "main.go", Line: 42, Function: "main.main"}
	attr := ReplaceAttr(nil, slog.Any(slog.SourceKey, source))
	if attr.Key != "logging.googleapis.com/sourceLocation" {
		t.Errorf("expected key 'logging.googleapis.com/sourceLocation', got %q", attr.Key)
	}
	if attr.Value.Kind() != slog.KindGroup {
		t.Errorf("expected group kind, got %v", attr.Value.Kind())
	}
}

func TestReplaceAttr_SourceKey_NonSource(t *testing.T) {
	attr := ReplaceAttr(nil, slog.String(slog.SourceKey, "not-a-source"))
	if attr.Key != slog.SourceKey {
		t.Errorf("expected key %q unchanged, got %q", slog.SourceKey, attr.Key)
	}
}

func TestReplaceAttr_WithGroups(t *testing.T) {
	attr := ReplaceAttr([]string{"group1"}, slog.String(slog.TimeKey, "2024-01-01"))
	// With groups, the attr should pass through unchanged.
	if attr.Key != slog.TimeKey {
		t.Errorf("expected key %q unchanged when groups present, got %q", slog.TimeKey, attr.Key)
	}
}

func TestReplaceAttr_UnknownKey(t *testing.T) {
	attr := ReplaceAttr(nil, slog.String("custom", "value"))
	if attr.Key != "custom" {
		t.Errorf("expected key 'custom' unchanged, got %q", attr.Key)
	}
	if attr.Value.String() != "value" {
		t.Errorf("expected value 'value', got %q", attr.Value.String())
	}
}
