package cmp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func CompareErr(t *testing.T, got error, want error, opts ...cmp.Option) {
	t.Helper()

	if want == nil {
		if got != nil {
			t.Fatalf("expected no error, got %T: %v", got, got)
		}
		return
	}

	if got == nil {
		t.Fatalf("expected %T error, got nil", want)
	}

	wantType := reflect.TypeOf(want)
	target := reflect.New(wantType)
	if !errors.As(got, target.Interface()) {
		t.Fatalf("expected error assignable to %v, got %T: %v", wantType, got, got)
	}

	typedGot := target.Elem().Interface().(error)
	if diff := cmp.Diff(want, typedGot, opts...); diff != "" {
		t.Errorf("error mismatch (-expected +got):\n%s", diff)
	}
}
