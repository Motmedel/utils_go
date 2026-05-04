package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func TestMakeGzipData_RoundTrip(t *testing.T) {
	t.Parallel()

	original := []byte("hello world hello world hello world")

	compressed, err := MakeGzipData(t.Context(), original)
	if err != nil {
		t.Fatalf("MakeGzipData error: %v", err)
	}

	if len(compressed) == 0 {
		t.Fatal("expected non-empty compressed output")
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll error: %v", err)
	}

	if !bytes.Equal(got, original) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, original)
	}
}

func TestMakeGzipData_Empty(t *testing.T) {
	t.Parallel()

	compressed, err := MakeGzipData(t.Context(), nil)
	if err != nil {
		t.Fatalf("MakeGzipData error: %v", err)
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}
	defer reader.Close()

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected empty output, got %d bytes", len(got))
	}
}

func TestMakeGzipData_Compresses(t *testing.T) {
	t.Parallel()

	original := bytes.Repeat([]byte("a"), 10000)

	compressed, err := MakeGzipData(t.Context(), original)
	if err != nil {
		t.Fatalf("MakeGzipData error: %v", err)
	}

	if len(compressed) >= len(original) {
		t.Fatalf("expected compressed (%d) smaller than original (%d)", len(compressed), len(original))
	}
}
