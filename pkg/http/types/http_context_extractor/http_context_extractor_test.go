package http_context_extractor

import (
	"context"
	"log/slog"
	"testing"
	"time"

	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

func newRecord() slog.Record {
	return slog.NewRecord(time.Now(), slog.LevelInfo, "message", 0)
}

// findAttr walks the (possibly nested-group) attributes of a record looking for
// a leaf attribute reachable via the given dotted path.
func findAttr(record *slog.Record, path ...string) (slog.Value, bool) {
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	return findInAttrs(attrs, path)
}

func findInAttrs(attrs []slog.Attr, path []string) (slog.Value, bool) {
	if len(path) == 0 {
		return slog.Value{}, false
	}

	for _, attr := range attrs {
		if attr.Key != path[0] {
			continue
		}
		if len(path) == 1 {
			return attr.Value, true
		}
		if attr.Value.Kind() == slog.KindGroup {
			if value, ok := findInAttrs(attr.Value.Group(), path[1:]); ok {
				return value, true
			}
		}
	}

	return slog.Value{}, false
}

func TestNew(t *testing.T) {
	t.Parallel()

	if New() == nil {
		t.Error("New() returned nil")
	}
}

func TestHandleNilRecord(t *testing.T) {
	t.Parallel()

	if err := New().Handle(context.Background(), nil); err != nil {
		t.Errorf("Handle(ctx, nil) = %v, want nil", err)
	}
}

func TestHandleEmptyContext(t *testing.T) {
	t.Parallel()

	record := newRecord()
	if err := New().Handle(context.Background(), &record); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if record.NumAttrs() != 0 {
		t.Errorf("NumAttrs = %d, want 0 for an empty context", record.NumAttrs())
	}
}

func TestHandleRequestId(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), motmedelHttpContext.RequestIdContextKey, "req-123")

	record := newRecord()
	if err := New().Handle(ctx, &record); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	value, ok := findAttr(&record, "http", "request", "id")
	if !ok {
		t.Fatal("http.request.id attribute not found")
	}
	if value.String() != "req-123" {
		t.Errorf("http.request.id = %q, want %q", value.String(), "req-123")
	}
}

func TestHandleRequestIdWrongType(t *testing.T) {
	t.Parallel()

	// A non-string value must be ignored rather than added.
	ctx := context.WithValue(context.Background(), motmedelHttpContext.RequestIdContextKey, 42)

	record := newRecord()
	if err := New().Handle(ctx, &record); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if _, ok := findAttr(&record, "http"); ok {
		t.Error("http attribute added for non-string request id, want none")
	}
}

func TestHandleHttpContext(t *testing.T) {
	t.Parallel()

	ctx := motmedelHttpContext.WithHttpContextValue(
		context.Background(),
		&motmedelHttpTypes.HttpContext{},
	)

	record := newRecord()
	if err := New().Handle(ctx, &record); err != nil {
		t.Fatalf("Handle: %v", err)
	}
}
