package cbor_schema_body_parser

import "testing"

func TestParse_NilSchema(t *testing.T) {
	t.Parallel()

	_, responseError := (&Parser[*testOrder]{}).Parse(nil, []byte{0xa0})
	if responseError == nil || responseError.ServerError == nil {
		t.Fatalf("expected a server error for a nil schema, got %#v", responseError)
	}
}

func TestNewWithSchema_NilSchema(t *testing.T) {
	t.Parallel()

	if _, err := NewWithSchema[*testOrder](nil); err == nil {
		t.Fatal("expected an error for a nil schema")
	}
}

func TestNew_UnsupportedType(t *testing.T) {
	t.Parallel()

	if _, err := New[chan int](); err == nil {
		t.Fatal("expected an error for a type without a CBOR schema")
	}
}
