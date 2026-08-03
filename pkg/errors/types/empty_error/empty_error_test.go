package empty_error

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
			err:      &Error{Field: "value"},
			expected: "empty value",
		},
		{
			name:     "with instance",
			err:      &Error{Field: "value", Instance: "request body"},
			expected: "empty value (request body)",
		},
		{
			name:     "empty field",
			err:      &Error{},
			expected: "empty ",
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

	err := New("value")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Field != "value" {
		t.Fatalf("expected Field %q, got %q", "value", err.Field)
	}
	if err.Instance != "" {
		t.Fatalf("expected empty Instance, got %q", err.Instance)
	}
}

func TestNewWithInstance(t *testing.T) {
	t.Parallel()

	err := NewWithInstance("value", "request body")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Field != "value" {
		t.Fatalf("expected Field %q, got %q", "value", err.Field)
	}
	if err.Instance != "request body" {
		t.Fatalf("expected Instance %q, got %q", "request body", err.Instance)
	}
}

func TestErrorInterface(t *testing.T) {
	t.Parallel()

	var _ error = (*Error)(nil)

	sentinel := New("value")
	wrapped := fmt.Errorf("wrap: %w", sentinel)

	if _, ok := errors.AsType[*Error](wrapped); !ok {
		t.Fatal("expected to extract *Error from wrapped error")
	}
}
