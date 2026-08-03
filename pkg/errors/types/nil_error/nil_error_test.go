package nil_error

import (
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "without instance",
			err:      &Error{Field: "pointer"},
			expected: "nil pointer",
		},
		{
			name:     "with instance",
			err:      &Error{Field: "pointer", Instance: "handler"},
			expected: "nil pointer (handler)",
		},
		{
			name:     "empty field",
			err:      &Error{},
			expected: "nil ",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.err.Error(); got != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, got)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	err := New("pointer")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Field != "pointer" {
		t.Fatalf("expected Field %q, got %q", "pointer", err.Field)
	}
	if err.Instance != "" {
		t.Fatalf("expected empty Instance, got %q", err.Instance)
	}
}

func TestNewWithInstance(t *testing.T) {
	t.Parallel()

	err := NewWithInstance("pointer", "handler")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Field != "pointer" {
		t.Fatalf("expected Field %q, got %q", "pointer", err.Field)
	}
	if err.Instance != "handler" {
		t.Fatalf("expected Instance %q, got %q", "handler", err.Instance)
	}
}

func TestErrorInterface(t *testing.T) {
	t.Parallel()

	var _ error = (*Error)(nil)

	sentinel := New("pointer")
	wrapped := fmt.Errorf("wrap: %w", sentinel)

	if _, ok := errors.AsType[*Error](wrapped); !ok {
		t.Fatal("expected to extract *Error from wrapped error")
	}
}
