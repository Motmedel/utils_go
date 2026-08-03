package mismatch_error

import (
	"errors"
	"fmt"
	"reflect"
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
			name:     "field set without values",
			err:      &Error{Field: "sub"},
			expected: "sub mismatch",
		},
		{
			name:     "field set with values",
			err:      &Error{Field: "sub", Values: []any{"expected", "actual"}},
			expected: "sub mismatch",
		},
		{
			name:     "empty field",
			err:      &Error{},
			expected: " mismatch",
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

	testCases := []struct {
		name           string
		field          string
		values         []any
		expectedValues []any
	}{
		{
			name:           "with values",
			field:          "sub",
			values:         []any{"expected", 42},
			expectedValues: []any{"expected", 42},
		},
		{
			name:           "without values",
			field:          "sub",
			values:         nil,
			expectedValues: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := New(testCase.field, testCase.values...)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Field != testCase.field {
				t.Fatalf("expected Field %q, got %q", testCase.field, err.Field)
			}
			if len(testCase.expectedValues) == 0 {
				if len(err.Values) != 0 {
					t.Fatalf("expected no values, got %v", err.Values)
				}
			} else if !reflect.DeepEqual(err.Values, testCase.expectedValues) {
				t.Fatalf("expected values %v, got %v", testCase.expectedValues, err.Values)
			}
		})
	}
}

func TestErrorInterface(t *testing.T) {
	t.Parallel()

	var _ error = (*Error)(nil)

	sentinel := New("sub", "a", "b")
	wrapped := fmt.Errorf("wrap: %w", sentinel)

	extracted, ok := errors.AsType[*Error](wrapped)
	if !ok {
		t.Fatal("expected to extract *Error from wrapped error")
	}
	if extracted.Field != "sub" {
		t.Fatalf("expected Field %q, got %q", "sub", extracted.Field)
	}
}
