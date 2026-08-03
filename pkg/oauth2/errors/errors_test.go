package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

var errUnrelated = errors.New("unrelated")

func TestRetrieveErrorIs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		target error
		want   bool
	}{
		{name: "matches ErrRetrieveToken", target: ErrRetrieveToken, want: true},
		{name: "does not match unrelated error", target: errUnrelated, want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := &RetrieveError{StatusCode: 400}
			if got := errors.Is(err, testCase.target); got != testCase.want {
				t.Errorf("errors.Is(%v, %v) = %t, want %t", err, testCase.target, got, testCase.want)
			}
		})
	}
}

func TestRetrieveErrorIsThroughWrapping(t *testing.T) {
	t.Parallel()

	err := &RetrieveError{StatusCode: http.StatusUnauthorized, ErrorCode: "invalid_client"}
	wrapped := fmt.Errorf("retrieve token: %w", err)

	if !errors.Is(wrapped, ErrRetrieveToken) {
		t.Errorf("errors.Is(wrapped, ErrRetrieveToken) = false, want true")
	}

	var extracted *RetrieveError
	if got, ok := errors.AsType[*RetrieveError](wrapped); !ok {
		t.Fatalf("errors.AsType[*RetrieveError] ok = false, want true")
	} else {
		extracted = got
	}

	if extracted.StatusCode != http.StatusUnauthorized {
		t.Errorf("extracted.StatusCode = %d, want %d", extracted.StatusCode, http.StatusUnauthorized)
	}
	if extracted.ErrorCode != "invalid_client" {
		t.Errorf("extracted.ErrorCode = %q, want %q", extracted.ErrorCode, "invalid_client")
	}
}

func TestRetrieveErrorError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  *RetrieveError
		want string
	}{
		{
			name: "status code only",
			err:  &RetrieveError{StatusCode: 500},
			want: "500",
		},
		{
			name: "error code only",
			err:  &RetrieveError{StatusCode: 400, ErrorCode: "invalid_grant"},
			want: `"invalid_grant"`,
		},
		{
			name: "error code and description",
			err:  &RetrieveError{ErrorCode: "invalid_grant", ErrorDescription: "bad code"},
			want: `"invalid_grant" "bad code"`,
		},
		{
			name: "error code description and uri",
			err: &RetrieveError{
				ErrorCode:        "invalid_grant",
				ErrorDescription: "bad code",
				ErrorURI:         "https://example.test/err",
			},
			want: `"invalid_grant" "bad code" "https://example.test/err"`,
		},
		{
			name: "error code and uri without description",
			err: &RetrieveError{
				ErrorCode: "invalid_grant",
				ErrorURI:  "https://example.test/err",
			},
			want: `"invalid_grant" "https://example.test/err"`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.err.Error(); got != testCase.want {
				t.Errorf("Error() = %q, want %q", got, testCase.want)
			}
		})
	}
}
