package token

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestTokenType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		tokenType string
		want      string
	}{
		{name: "explicit type", tokenType: "MAC", want: "MAC"},
		{name: "empty falls back to Bearer", tokenType: "", want: "Bearer"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tok := &Token{AccessToken: "abc", TokenType: testCase.tokenType}
			if got := tok.Type(); got != testCase.want {
				t.Errorf("Type() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestTokenSetAuthHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		tokenType string
		access    string
		want      string
	}{
		{name: "bearer default", tokenType: "", access: "abc123", want: "Bearer abc123"},
		{name: "explicit type", tokenType: "MAC", access: "xyz", want: "MAC xyz"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test", nil)
			if err != nil {
				t.Fatalf("http.NewRequestWithContext: %v", err)
			}

			tok := &Token{AccessToken: testCase.access, TokenType: testCase.tokenType}
			tok.SetAuthHeader(req)

			if got := req.Header.Get("Authorization"); got != testCase.want {
				t.Errorf("Authorization header = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestTokenExpiredAndValid(t *testing.T) {
	t.Parallel()

	// The expiry logic reads the wall clock, so anchor cases far from `now`
	// (well beyond the 10s expiryDelta) to keep results deterministic.
	now := time.Now()

	testCases := []struct {
		name        string
		accessToken string
		expiry      time.Time
		wantExpired bool
		wantValid   bool
	}{
		{
			name:        "zero expiry never expires",
			accessToken: "abc",
			expiry:      time.Time{},
			wantExpired: false,
			wantValid:   true,
		},
		{
			name:        "future expiry not expired",
			accessToken: "abc",
			expiry:      now.Add(time.Hour),
			wantExpired: false,
			wantValid:   true,
		},
		{
			name:        "past expiry expired",
			accessToken: "abc",
			expiry:      now.Add(-time.Hour),
			wantExpired: true,
			wantValid:   false,
		},
		{
			name:        "within delta counts as expired",
			accessToken: "abc",
			expiry:      now.Add(5 * time.Second),
			wantExpired: true,
			wantValid:   false,
		},
		{
			name:        "empty access token is invalid",
			accessToken: "",
			expiry:      now.Add(time.Hour),
			wantExpired: false,
			wantValid:   false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tok := &Token{AccessToken: testCase.accessToken, Expiry: testCase.expiry}
			if got := tok.expired(); got != testCase.wantExpired {
				t.Errorf("expired() = %t, want %t", got, testCase.wantExpired)
			}
			if got := tok.Valid(); got != testCase.wantValid {
				t.Errorf("Valid() = %t, want %t", got, testCase.wantValid)
			}
		})
	}
}

func TestTokenValidNil(t *testing.T) {
	t.Parallel()

	var tok *Token
	if tok.Valid() {
		t.Errorf("(*Token)(nil).Valid() = true, want false")
	}
}

func TestTokenExtra(t *testing.T) {
	t.Parallel()

	t.Run("nil raw", func(t *testing.T) {
		t.Parallel()

		tok := &Token{AccessToken: "abc"}
		if got := tok.Extra("scope"); got != nil {
			t.Errorf("Extra on nil Raw = %v, want nil", got)
		}
	})

	t.Run("present key", func(t *testing.T) {
		t.Parallel()

		tok := &Token{Raw: map[string]any{"scope": "read write", "count": 3}}
		if got := tok.Extra("scope"); got != "read write" {
			t.Errorf("Extra(scope) = %v, want %q", got, "read write")
		}
		if got := tok.Extra("count"); got != 3 {
			t.Errorf("Extra(count) = %v, want %d", got, 3)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		t.Parallel()

		tok := &Token{Raw: map[string]any{"scope": "read"}}
		if got := tok.Extra("missing"); got != nil {
			t.Errorf("Extra(missing) = %v, want nil", got)
		}
	})
}

func TestTokenMarshalRoundTrip(t *testing.T) {
	t.Parallel()

	expiry := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)

	testCases := []struct {
		name string
		tok  *Token
	}{
		{
			name: "full token",
			tok: &Token{
				AccessToken:  "access",
				TokenType:    "Bearer",
				RefreshToken: "refresh",
				Expiry:       expiry,
				ExpiresIn:    3600,
			},
		},
		{
			name: "minimal token",
			tok:  &Token{AccessToken: "access"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // G117: deliberately marshaling the token to verify its JSON round-trip.
			data, err := json.Marshal(testCase.tok)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			var got Token
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("json.Unmarshal: %v", err)
			}

			if got.AccessToken != testCase.tok.AccessToken {
				t.Errorf("AccessToken = %q, want %q", got.AccessToken, testCase.tok.AccessToken)
			}
			if got.TokenType != testCase.tok.TokenType {
				t.Errorf("TokenType = %q, want %q", got.TokenType, testCase.tok.TokenType)
			}
			if got.RefreshToken != testCase.tok.RefreshToken {
				t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, testCase.tok.RefreshToken)
			}
			if !got.Expiry.Equal(testCase.tok.Expiry) {
				t.Errorf("Expiry = %v, want %v", got.Expiry, testCase.tok.Expiry)
			}
			if got.ExpiresIn != testCase.tok.ExpiresIn {
				t.Errorf("ExpiresIn = %d, want %d", got.ExpiresIn, testCase.tok.ExpiresIn)
			}
		})
	}
}

func TestTokenUnmarshalFromResponseJSON(t *testing.T) {
	t.Parallel()

	data := []byte(`{"access_token":"at","token_type":"Bearer","refresh_token":"rt","expires_in":7200}`)

	var tok Token
	if err := json.Unmarshal(data, &tok); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if tok.AccessToken != "at" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "at")
	}
	if tok.TokenType != "Bearer" {
		t.Errorf("TokenType = %q, want %q", tok.TokenType, "Bearer")
	}
	if tok.RefreshToken != "rt" {
		t.Errorf("RefreshToken = %q, want %q", tok.RefreshToken, "rt")
	}
	if tok.ExpiresIn != 7200 {
		t.Errorf("ExpiresIn = %d, want %d", tok.ExpiresIn, 7200)
	}
}
