package errors

import (
	"errors"
	"fmt"
	"testing"
)

var (
	errBoom     = errors.New("boom")
	errSentinel = errors.New("sentinel")
)

func TestNon2xxStatusCodeError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		statusCode   int
		expectedCode string
	}{
		{name: "zero status code", statusCode: 0, expectedCode: ""},
		{name: "not found", statusCode: 404, expectedCode: "404"},
		{name: "internal server error", statusCode: 500, expectedCode: "500"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := &Non2xxStatusCodeError{StatusCode: testCase.statusCode}

			if err.Error() != ErrNon2xxStatusCode.Error() {
				t.Errorf("Error() = %q, want %q", err.Error(), ErrNon2xxStatusCode.Error())
			}
			if !errors.Is(err, ErrNon2xxStatusCode) {
				t.Error("errors.Is(err, ErrNon2xxStatusCode) = false, want true")
			}
			if got := err.GetInput(); got != testCase.statusCode {
				t.Errorf("GetInput() = %v, want %v", got, testCase.statusCode)
			}
			if got := err.GetCode(); got != testCase.expectedCode {
				t.Errorf("GetCode() = %q, want %q", got, testCase.expectedCode)
			}
		})
	}
}

func TestNon2xxStatusCodeErrorIsFalse(t *testing.T) {
	t.Parallel()

	err := &Non2xxStatusCodeError{StatusCode: 500}
	if errors.Is(err, ErrReattemptFailedError) {
		t.Error("errors.Is(err, ErrReattemptFailedError) = true, want false")
	}
}

func TestReattemptFailedError(t *testing.T) {
	t.Parallel()

	cause := errBoom
	err := &ReattemptFailedError{Attempt: 3, Cause: cause}

	if err.Error() != ErrReattemptFailedError.Error() {
		t.Errorf("Error() = %q, want %q", err.Error(), ErrReattemptFailedError.Error())
	}
	if !errors.Is(err, ErrReattemptFailedError) {
		t.Error("errors.Is(err, ErrReattemptFailedError) = false, want true")
	}
	if !errors.Is(err.GetCause(), cause) {
		t.Errorf("GetCause() = %v, want %v", err.GetCause(), cause)
	}
	if !errors.Is(err.Unwrap(), cause) {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is(err, cause) = false, want true")
	}
}

func TestReattemptFailedErrorWrappedCause(t *testing.T) {
	t.Parallel()

	sentinel := errSentinel
	err := &ReattemptFailedError{Attempt: 1, Cause: fmt.Errorf("wrapping: %w", sentinel)}

	if !errors.Is(err, sentinel) {
		t.Error("errors.Is(err, sentinel) = false, want true")
	}
}
