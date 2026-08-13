package brotli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestMakeBrotliData(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "short text", data: []byte("hello")},
		{name: "compressible text", data: []byte(strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100))},
		{name: "binary", data: bytes.Repeat([]byte{0x00, 0xff, 0x10, 0x20}, 256)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			compressed, err := MakeBrotliData(context.Background(), testCase.data)
			if err != nil {
				t.Fatalf("make brotli data: %v", err)
			}

			decompressed, err := io.ReadAll(NewReader(bytes.NewReader(compressed)))
			if err != nil {
				t.Fatalf("read all: %v", err)
			}
			if !bytes.Equal(decompressed, testCase.data) {
				t.Error("expected the decompressed data to equal the input")
			}
		})
	}
}

func TestMakeBrotliDataBeatsSize(t *testing.T) {
	t.Parallel()
	data := []byte(strings.Repeat("<div class=\"item\">repetitive markup</div>", 200))

	compressed, err := MakeBrotliData(context.Background(), data)
	if err != nil {
		t.Fatalf("make brotli data: %v", err)
	}
	if len(compressed) >= len(data) {
		t.Errorf("expected compression to reduce %d bytes, got %d", len(data), len(compressed))
	}
}
