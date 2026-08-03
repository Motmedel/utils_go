package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"testing"
)

func dropTime(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 && attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}

// newTreeLogger returns a logger whose handler is the tree Handler wrapping a
// JSON handler that writes to buf (with the time attribute stripped).
func newTreeLogger(buf *bytes.Buffer, opts *slog.HandlerOptions) *slog.Logger {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	opts.ReplaceAttr = dropTime
	return slog.New(New(slog.NewJSONHandler(buf, opts)))
}

func parseLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &m); err != nil {
		t.Fatalf("failed to parse log output %q: %v", buf.String(), err)
	}
	delete(m, slog.LevelKey)
	delete(m, slog.MessageKey)
	return m
}

func TestHandleBasicAttrs(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	newTreeLogger(buf, nil).Info("hello", "key", "value")

	got := parseLog(t, buf)
	if want := (map[string]any{"key": "value"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestHandleMergesSameGroup(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := newTreeLogger(buf, nil).With(slog.Group("a", slog.Int("x", 1)))
	logger.Info("msg", slog.Group("a", slog.Int("y", 2)))

	got := parseLog(t, buf)
	want := map[string]any{"a": map[string]any{"x": float64(1), "y": float64(2)}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestHandleWithGroupNesting(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := newTreeLogger(buf, nil).WithGroup("g").With("x", 1)
	logger.Info("msg", "y", 2)

	got := parseLog(t, buf)
	want := map[string]any{"g": map[string]any{"x": float64(1), "y": float64(2)}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestHandleOverridesDuplicateKey(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := newTreeLogger(buf, nil).With("a", 1)
	logger.Info("msg", "a", 2)

	got := parseLog(t, buf)
	if want := (map[string]any{"a": float64(2)}); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestHandleNoAttrsRendersRoot(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := newTreeLogger(buf, nil).With("persistent", "yes")
	logger.Info("msg")

	got := parseLog(t, buf)
	if want := (map[string]any{"persistent": "yes"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestWithAttrsDoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	base := newTreeLogger(buf, nil)
	_ = base.With("derived", "value")

	base.Info("msg", "own", "x")

	got := parseLog(t, buf)
	if _, ok := got["derived"]; ok {
		t.Fatalf("derived attr leaked into base logger: %#v", got)
	}
	if want := (map[string]any{"own": "x"}); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestEnabledDelegatesToNext(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	h := New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("expected Info to be disabled at Warn level")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Fatal("expected Error to be enabled at Warn level")
	}
}
