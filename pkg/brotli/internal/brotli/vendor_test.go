package brotli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// The vendored implementation is covered thoroughly upstream; this exercises
// the round trip used by the facade.
func TestVendoredRoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "text", data: []byte(strings.Repeat("vendored round trip ", 100))},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var buffer bytes.Buffer
			writer := NewWriterLevel(&buffer, BestCompression)
			if _, err := writer.Write(testCase.data); err != nil {
				t.Fatalf("write: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}

			decompressed, err := io.ReadAll(NewReader(&buffer))
			if err != nil {
				t.Fatalf("read all: %v", err)
			}
			if !bytes.Equal(decompressed, testCase.data) {
				t.Error("expected the decompressed data to equal the input")
			}
		})
	}
}
