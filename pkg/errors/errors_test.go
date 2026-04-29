package errors

import (
	"fmt"
	"testing"
)

// uncomparableError contains a slice field, making it uncomparable with ==.
type uncomparableError struct {
	Details []string
}

func (e uncomparableError) Error() string {
	return fmt.Sprintf("uncomparable: %v", e.Details)
}

func (e uncomparableError) Unwrap() error {
	return fmt.Errorf("inner error")
}

func TestCollectWrappedErrors_UncomparableType(t *testing.T) {
	err := uncomparableError{Details: []string{"a", "b"}}

	// This must not panic.
	results := CollectWrappedErrors(err)

	if len(results) != 1 {
		t.Fatalf("expected 1 wrapped error, got %d", len(results))
	}
	if results[0].Error() != "inner error" {
		t.Fatalf("expected 'inner error', got %q", results[0].Error())
	}
}

func TestCollectWrappedErrors_ComparableType(t *testing.T) {
	inner := fmt.Errorf("inner")
	err := fmt.Errorf("outer: %w", inner)

	results := CollectWrappedErrors(err)

	if len(results) != 1 {
		t.Fatalf("expected 1 wrapped error, got %d", len(results))
	}
	if results[0] != inner {
		t.Fatalf("expected inner error, got %v", results[0])
	}
}

func TestCollectWrappedErrors_NilError(t *testing.T) {
	results := CollectWrappedErrors(nil)

	if len(results) != 0 {
		t.Fatalf("expected 0 wrapped errors, got %d", len(results))
	}
}

// Ensure structurally identical wrapped errors are NOT skipped.
func TestCollectWrappedErrors_StructurallyIdenticalChild(t *testing.T) {
	// Use an uncomparable error without Unwrap as the child.
	child := uncomparableLeafError{Details: []string{"a", "b"}}
	parent := wrappingError{msg: "parent", wrapped: child}

	results := CollectWrappedErrors(parent)

	// child is structurally identical to what root would look like,
	// but it's a different instance — it must still be collected.
	if len(results) != 1 {
		t.Fatalf("expected 1 wrapped error, got %d", len(results))
	}
}

// Verify that reflect.DeepEqual would incorrectly skip this case.
func TestCollectWrappedErrors_DeepEqualWouldSkip(t *testing.T) {
	// Parent and child are structurally identical uncomparable errors.
	child := uncomparableLeafError{Details: []string{"x"}}
	parent := uncomparableLeafError{Details: []string{"x"}}
	parent.wrapped = child

	results := CollectWrappedErrors(parent)

	// reflect.DeepEqual would consider child == parent and skip it.
	// Our fix must still collect the child.
	if len(results) != 1 {
		t.Fatalf("expected 1 wrapped error, got %d", len(results))
	}
}

type wrappingError struct {
	msg     string
	wrapped error
}

func (e wrappingError) Error() string { return e.msg }
func (e wrappingError) Unwrap() error { return e.wrapped }

// uncomparableLeafError is uncomparable (has a slice) and optionally wraps another error.
type uncomparableLeafError struct {
	Details []string
	wrapped error
}

func (e uncomparableLeafError) Error() string {
	return fmt.Sprintf("leaf: %v", e.Details)
}

func (e uncomparableLeafError) Unwrap() error { return e.wrapped }
