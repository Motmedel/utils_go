package json

import (
	"errors"
	"strings"
	"testing"
)

type sample struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// failingReader always returns an error from Read.
type failingReader struct{}

var errRead = errors.New("read failure")

func (failingReader) Read([]byte) (int, error) {
	return 0, errRead
}

func TestDecodeJson(t *testing.T) {
	t.Parallel()

	t.Run("valid object", func(t *testing.T) {
		t.Parallel()

		got, err := DecodeJson[sample](strings.NewReader(`{"name":"alice","count":3}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "alice" || got.Count != 3 {
			t.Fatalf("expected {alice 3}, got %+v", got)
		}
	})

	t.Run("round-trip via ObjectToBytes", func(t *testing.T) {
		t.Parallel()

		data, err := ObjectToBytes(&sample{Name: "bob", Count: 7})
		if err != nil {
			t.Fatalf("unexpected error marshalling: %v", err)
		}
		got, err := DecodeJson[sample](strings.NewReader(string(data)))
		if err != nil {
			t.Fatalf("unexpected error decoding: %v", err)
		}
		if got.Name != "bob" || got.Count != 7 {
			t.Fatalf("expected {bob 7}, got %+v", got)
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		t.Parallel()

		_, err := DecodeJson[sample](strings.NewReader(`{"name":`))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("reader error is propagated", func(t *testing.T) {
		t.Parallel()

		_, err := DecodeJson[sample](failingReader{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, errRead) {
			t.Fatalf("expected wrapped read error, got %v", err)
		}
	})

	t.Run("map target", func(t *testing.T) {
		t.Parallel()

		got, err := DecodeJson[map[string]any](strings.NewReader(`{"a":1,"b":"x"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 keys, got %d (%v)", len(got), got)
		}
	})
}

func TestObjectToMap(t *testing.T) {
	t.Parallel()

	t.Run("struct converted to map", func(t *testing.T) {
		t.Parallel()

		got, err := ObjectToMap(&sample{Name: "carol", Count: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got["name"] != "carol" {
			t.Fatalf("expected name=carol, got %v", got["name"])
		}
		// JSON numbers decode into float64.
		if count, ok := got["count"].(float64); !ok || count != 5 {
			t.Fatalf("expected count=5, got %v (%T)", got["count"], got["count"])
		}
	})

	t.Run("marshal error for unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ObjectToMap(make(chan int))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("non-object json fails to unmarshal into map", func(t *testing.T) {
		t.Parallel()

		// Marshals to the JSON number 5, which cannot unmarshal into a map.
		_, err := ObjectToMap(5)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestObjectToBytes(t *testing.T) {
	t.Parallel()

	t.Run("struct converted to bytes", func(t *testing.T) {
		t.Parallel()

		got, err := ObjectToBytes(&sample{Name: "dave", Count: 9})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Round-trip back into a map to assert content independent of key ordering.
		roundTripped, err := DecodeJson[map[string]any](strings.NewReader(string(got)))
		if err != nil {
			t.Fatalf("unexpected error decoding: %v", err)
		}
		if roundTripped["name"] != "dave" {
			t.Fatalf("expected name=dave, got %v", roundTripped["name"])
		}
	})

	t.Run("marshal error for unsupported type", func(t *testing.T) {
		t.Parallel()

		_, err := ObjectToBytes(make(chan int))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
