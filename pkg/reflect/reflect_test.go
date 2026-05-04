package reflect

import (
	"reflect"
	"testing"
)

type sample struct{}

type generic[T any] struct{}

func TestTypeOf_Basic(t *testing.T) {
	t.Parallel()

	got := TypeOf[int]()
	if got.Kind() != reflect.Int {
		t.Fatalf("expected int, got %v", got)
	}

	if name := TypeOf[sample]().Name(); name != "sample" {
		t.Fatalf("expected sample, got %q", name)
	}
}

func TestTypeOf_Interface(t *testing.T) {
	t.Parallel()
	got := TypeOf[error]()
	if got.Kind() != reflect.Interface {
		t.Fatalf("expected interface kind, got %v", got)
	}
}

func TestRemoveIndirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   reflect.Type
		want reflect.Kind
	}{
		{"value", reflect.TypeOf(1), reflect.Int},
		{"single-pointer", reflect.TypeOf((*int)(nil)), reflect.Int},
		{"double-pointer", reflect.TypeOf((**int)(nil)), reflect.Int},
		{"struct-pointer", reflect.TypeOf((*sample)(nil)), reflect.Struct},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := RemoveIndirection(tt.in)
			if got.Kind() != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got.Kind())
			}
		})
	}
}

func TestGetTypeName_NonGeneric(t *testing.T) {
	t.Parallel()
	name, isGeneric := GetTypeName(reflect.TypeOf(sample{}))
	if isGeneric {
		t.Fatal("expected non-generic")
	}
	if name != "sample" {
		t.Fatalf("got %q", name)
	}
}

func TestGetTypeName_Generic(t *testing.T) {
	t.Parallel()
	name, isGeneric := GetTypeName(reflect.TypeOf(generic[int]{}))
	if !isGeneric {
		t.Fatal("expected generic")
	}
	if name != "generic" {
		t.Fatalf("got %q", name)
	}
}
