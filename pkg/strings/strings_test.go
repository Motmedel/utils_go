package strings

import (
	"bytes"
	"testing"
	"time"
)

func TestByteSliceFromAny_PlainBytes(t *testing.T) {
	t.Parallel()
	in := []byte("hello")
	got, ok := ByteSliceFromAny(in)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !bytes.Equal(got, in) {
		t.Fatalf("expected %q, got %q", in, got)
	}
}

func TestByteSliceFromAny_NamedByteSlice(t *testing.T) {
	t.Parallel()
	type Bytes []byte
	in := Bytes("hello")
	got, ok := ByteSliceFromAny(in)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("got %q", got)
	}
}

func TestByteSliceFromAny_NotByteSlice(t *testing.T) {
	t.Parallel()
	if _, ok := ByteSliceFromAny("hello"); ok {
		t.Fatal("expected ok=false for string input")
	}
	if _, ok := ByteSliceFromAny(123); ok {
		t.Fatal("expected ok=false for int input")
	}
	if _, ok := ByteSliceFromAny(nil); ok {
		t.Fatal("expected ok=false for nil input")
	}
}

func TestMakeTextualRepresentation_String(t *testing.T) {
	t.Parallel()
	got, err := MakeTextualRepresentation("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "abc" {
		t.Fatalf("expected %q, got %q", "abc", got)
	}
}

func TestMakeTextualRepresentation_Time(t *testing.T) {
	t.Parallel()
	tm := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	got, err := MakeTextualRepresentation(tm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "2025-01-02T03:04:05Z" {
		t.Fatalf("got %q", got)
	}
}

type textMarshaler struct {
	value string
}

func (m *textMarshaler) MarshalText() ([]byte, error) {
	return []byte(m.value), nil
}

func TestMakeTextualRepresentation_TextMarshaler(t *testing.T) {
	t.Parallel()
	got, err := MakeTextualRepresentation(&textMarshaler{value: "marshalled"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "marshalled" {
		t.Fatalf("got %q", got)
	}
}

func TestMakeTextualRepresentation_ByteSlice(t *testing.T) {
	t.Parallel()
	got, err := MakeTextualRepresentation([]byte("raw"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "raw" {
		t.Fatalf("got %q", got)
	}
}

func TestMakeTextualRepresentation_Fallback(t *testing.T) {
	t.Parallel()
	got, err := MakeTextualRepresentation(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "42" {
		t.Fatalf("got %q", got)
	}
}

func TestShellJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want string
	}{
		{"empty", nil, ""},
		{"safe", []string{"foo", "bar"}, "foo bar"},
		{"empty-string", []string{""}, "''"},
		{"with-space", []string{"hello world"}, "'hello world'"},
		{"with-quote", []string{"it's"}, "'it'\"'\"'s'"},
		{"safe-chars", []string{"a/b.c-d_e@f"}, "a/b.c-d_e@f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ShellJoin(tt.in); got != tt.want {
				t.Fatalf("ShellJoin(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestHasAnyPrefix(t *testing.T) {
	t.Parallel()

	if !HasAnyPrefix("hello world", "hi", "hello") {
		t.Fatal("expected true")
	}
	if HasAnyPrefix("hello world", "hi", "world") {
		t.Fatal("expected false")
	}
	if HasAnyPrefix("hello") {
		t.Fatal("expected false with no prefixes")
	}
}
