package log

import (
	"log/slog"
	"testing"
	"time"
)

func TestReplaceAttr(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		groups  []string
		attr    slog.Attr
		wantKey string
		// assert inspects the returned attr beyond its key.
		assert func(t *testing.T, got slog.Attr)
	}{
		{
			name:    "non-empty groups returns attr unchanged",
			groups:  []string{"nested"},
			attr:    slog.Time(slog.TimeKey, testTime()),
			wantKey: slog.TimeKey,
		},
		{
			name:    "time key is renamed",
			groups:  nil,
			attr:    slog.Time(slog.TimeKey, testTime()),
			wantKey: "@timestamp",
			assert: func(t *testing.T, got slog.Attr) {
				if got.Value.Kind() != slog.KindTime {
					t.Errorf("value kind = %v, want time (value should be preserved)", got.Value.Kind())
				}
			},
		},
		{
			name:    "message key is renamed",
			groups:  nil,
			attr:    slog.String(slog.MessageKey, "hello"),
			wantKey: "message",
			assert: func(t *testing.T, got slog.Attr) {
				if got.Value.String() != "hello" {
					t.Errorf("value = %q, want hello", got.Value.String())
				}
			},
		},
		{
			name:    "string level is grouped and lowercased",
			groups:  nil,
			attr:    slog.String(slog.LevelKey, "INFO"),
			wantKey: "log",
			assert: func(t *testing.T, got slog.Attr) {
				if got.Value.Kind() != slog.KindGroup {
					t.Fatalf("value kind = %v, want group", got.Value.Kind())
				}
				group := got.Value.Group()
				if len(group) != 1 {
					t.Fatalf("group length = %d, want 1 (%v)", len(group), group)
				}
				if group[0].Key != "level" {
					t.Errorf("group key = %q, want level", group[0].Key)
				}
				if group[0].Value.String() != "info" {
					t.Errorf("group level value = %q, want info", group[0].Value.String())
				}
			},
		},
		{
			name:    "mixed case string level is lowercased",
			groups:  nil,
			attr:    slog.String(slog.LevelKey, "Warn"),
			wantKey: "log",
			assert: func(t *testing.T, got slog.Attr) {
				group := got.Value.Group()
				if len(group) != 1 || group[0].Value.String() != "warn" {
					t.Errorf("group = %v, want single level=warn", group)
				}
			},
		},
		{
			name:    "non-string level value returns attr unchanged",
			groups:  nil,
			attr:    slog.Any(slog.LevelKey, slog.LevelInfo),
			wantKey: slog.LevelKey,
			assert: func(t *testing.T, got slog.Attr) {
				if got.Value.Kind() == slog.KindGroup {
					t.Error("non-string level should not be grouped")
				}
			},
		},
		{
			name:    "unknown key returns attr unchanged",
			groups:  nil,
			attr:    slog.String("custom", "value"),
			wantKey: "custom",
			assert: func(t *testing.T, got slog.Attr) {
				if got.Value.String() != "value" {
					t.Errorf("value = %q, want value", got.Value.String())
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := ReplaceAttr(testCase.groups, testCase.attr)
			if got.Key != testCase.wantKey {
				t.Errorf("ReplaceAttr key = %q, want %q", got.Key, testCase.wantKey)
			}
			if testCase.assert != nil {
				testCase.assert(t, got)
			}
		})
	}
}

func testTime() time.Time {
	return time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
}
