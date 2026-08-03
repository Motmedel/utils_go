package missing_error

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
			name:     "field set",
			err:      &Error{Field: "registry"},
			expected: "missing registry",
		},
		{
			name:     "empty field",
			err:      &Error{},
			expected: "missing ",
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

	err := New("registry")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Field != "registry" {
		t.Fatalf("expected Field %q, got %q", "registry", err.Field)
	}
}

func TestErrorInterface(t *testing.T) {
	t.Parallel()

	var _ error = (*Error)(nil)

	sentinel := New("registry")
	wrapped := fmt.Errorf("wrap: %w", sentinel)

	if _, ok := errors.AsType[*Error](wrapped); !ok {
		t.Fatal("expected to extract *Error from wrapped error")
	}
}
