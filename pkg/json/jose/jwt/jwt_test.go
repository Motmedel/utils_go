package jwt

import (
	"errors"
	"testing"
	"time"

	jwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
)

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
