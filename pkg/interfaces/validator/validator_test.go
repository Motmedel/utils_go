package validator

import (
	"errors"
	"testing"
)

var errSentinel = errors.New("validator failure")

// mockValidator is a minimal implementation of the Validator interface.
type mockValidator struct {
	err error
}

func (m *mockValidator) Validate(_ string) error {
	return m.err
}

// Compile-time assertions that the mock and the Function adapter satisfy the interface.
var (
	_ Validator[string] = (*mockValidator)(nil)
	_ Validator[string] = Function[string](nil)
)

func TestMockThroughInterface(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{name: "valid", err: nil, wantErr: false},
		{name: "invalid", err: errSentinel, wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var validator Validator[string] = &mockValidator{err: testCase.err}
			err := validator.Validate("input")
			if testCase.wantErr {
				if !errors.Is(err, errSentinel) {
					t.Fatalf("expected sentinel error, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	validator := New(func(input string) error {
		if len(input) < 3 {
			return errSentinel
		}
		return nil
	})

	if err := validator.Validate("ok"); !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if err := validator.Validate("long enough"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
