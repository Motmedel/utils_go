package jwt

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
)

func TestSplitToken(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		token       string
		expected    [3]string
		expectError bool
	}{
		{
			name:        "empty token",
			token:       "",
			expected:    [3]string{},
			expectError: true,
		},
		{
			name:        "single part",
			token:       "header",
			expected:    [3]string{},
			expectError: true,
		},
		{
			name:        "two parts",
			token:       "header.payload",
			expected:    [3]string{},
			expectError: true,
		},
		{
			name:        "three parts",
			token:       "header.payload.signature",
			expected:    [3]string{"header", "payload", "signature"},
			expectError: false,
		},
		{
			name:        "four parts (extra dots in signature)",
			token:       "header.payload.sig.nature",
			expected:    [3]string{"header", "payload", "sig.nature"},
			expectError: false,
		},
		{
			name:        "empty parts",
			token:       "..",
			expected:    [3]string{"", "", ""},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parts, err := SplitToken(tc.token)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if !errors.Is(err, motmedelErrors.ErrBadSplit) {
					t.Fatalf("expected ErrBadSplit, got: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
				if parts != tc.expected {
					t.Fatalf("expected %v, got %v", tc.expected, parts)
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	// Create valid base64-encoded parts
	validHeader := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
	validPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"1234567890"}`))
	validSignature := base64.RawURLEncoding.EncodeToString([]byte("signature"))

	testCases := []struct {
		name           string
		token          string
		expectedHeader []byte
		expectError    bool
	}{
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "invalid split",
			token:       "only.two",
			expectError: true,
		},
		{
			name:        "invalid base64 header",
			token:       "!!!invalid!!!" + "." + validPayload + "." + validSignature,
			expectError: true,
		},
		{
			name:        "invalid base64 payload",
			token:       validHeader + ".!!!invalid!!!." + validSignature,
			expectError: true,
		},
		{
			name:        "invalid base64 signature",
			token:       validHeader + "." + validPayload + ".!!!invalid!!!",
			expectError: true,
		},
		{
			name:           "valid token",
			token:          validHeader + "." + validPayload + "." + validSignature,
			expectedHeader: []byte(`{"alg":"HS256"}`),
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			header, _, _, err := Parse(tc.token)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
				if string(header) != string(tc.expectedHeader) {
					t.Fatalf("expected header %s, got %s", tc.expectedHeader, header)
				}
			}
		})
	}
}

func TestValidateExpiresAt(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name        string
		expiresAt   time.Time
		cmp         time.Time
		expectError error
	}{
		{
			name:        "not expired",
			expiresAt:   now.Add(time.Hour),
			cmp:         now,
			expectError: nil,
		},
		{
			name:        "expired",
			expiresAt:   now.Add(-time.Hour),
			cmp:         now,
			expectError: jwtErrors.ErrExpExpired,
		},
		{
			name:        "exactly at expiration",
			expiresAt:   now,
			cmp:         now,
			expectError: nil,
		},
		{
			name:        "one nanosecond after expiration",
			expiresAt:   now,
			cmp:         now.Add(time.Nanosecond),
			expectError: jwtErrors.ErrExpExpired,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateExpiresAt(tc.expiresAt, tc.cmp)

			if tc.expectError == nil {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
			} else {
				if !errors.Is(err, tc.expectError) {
					t.Fatalf("expected error %v, got: %v", tc.expectError, err)
				}
			}
		})
	}
}

func TestValidateNotBefore(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name        string
		notBefore   time.Time
		cmp         time.Time
		expectError error
	}{
		{
			name:        "after not before",
			notBefore:   now.Add(-time.Hour),
			cmp:         now,
			expectError: nil,
		},
		{
			name:        "before not before",
			notBefore:   now.Add(time.Hour),
			cmp:         now,
			expectError: jwtErrors.ErrNbfBefore,
		},
		{
			name:        "exactly at not before",
			notBefore:   now,
			cmp:         now,
			expectError: nil,
		},
		{
			name:        "one nanosecond before not before",
			notBefore:   now,
			cmp:         now.Add(-time.Nanosecond),
			expectError: jwtErrors.ErrNbfBefore,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateNotBefore(tc.notBefore, tc.cmp)

			if tc.expectError == nil {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
			} else {
				if !errors.Is(err, tc.expectError) {
					t.Fatalf("expected error %v, got: %v", tc.expectError, err)
				}
			}
		})
	}
}

func TestValidateIssuedAt(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name        string
		issuedAt    time.Time
		cmp         time.Time
		expectError error
	}{
		{
			name:        "after issued at",
			issuedAt:    now.Add(-time.Hour),
			cmp:         now,
			expectError: nil,
		},
		{
			name:        "before issued at",
			issuedAt:    now.Add(time.Hour),
			cmp:         now,
			expectError: jwtErrors.ErrIatBefore,
		},
		{
			name:        "exactly at issued at",
			issuedAt:    now,
			cmp:         now,
			expectError: nil,
		},
		{
			name:        "one nanosecond before issued at",
			issuedAt:    now,
			cmp:         now.Add(-time.Nanosecond),
			expectError: jwtErrors.ErrIatBefore,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateIssuedAt(tc.issuedAt, tc.cmp)

			if tc.expectError == nil {
				if err != nil {
					t.Fatalf("expected no error but got: %v", err)
				}
			} else {
				if !errors.Is(err, tc.expectError) {
					t.Fatalf("expected error %v, got: %v", tc.expectError, err)
				}
			}
		})
	}
}
