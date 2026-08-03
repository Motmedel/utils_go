package rate_limiting

import (
	"net/http"
	"testing"
	"testing/synctest"
	"time"
)

func TestClaim(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		rateLimiter := &RateLimiter{Bucket: make([]*time.Time, 2), NumSecondsExpiry: 5}

		wantExpiry := time.Now().Add(5 * time.Second)

		expiry1, full1 := rateLimiter.Claim()
		expiry2, full2 := rateLimiter.Claim()
		if full1 || full2 {
			t.Fatalf("claims within capacity must not be limited (full1=%v full2=%v)", full1, full2)
		}
		if expiry1 == nil || !expiry1.Equal(wantExpiry) {
			t.Fatalf("expiry1 = %v, want %v", expiry1, wantExpiry)
		}
		if expiry2 == nil || !expiry2.Equal(wantExpiry) {
			t.Fatalf("expiry2 = %v, want %v", expiry2, wantExpiry)
		}

		if _, full := rateLimiter.Claim(); !full {
			t.Fatal("a claim beyond capacity must be limited")
		}

		// Bucket slots are freed one second before expiry; advancing past the window
		// fires those timers.
		time.Sleep(5 * time.Second)
		synctest.Wait()

		if _, full := rateLimiter.Claim(); full {
			t.Fatal("a claim after the window elapsed must not be limited")
		}
	})
}

func TestDefaultGetRateLimitingKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		nilRequest bool
		remoteAddr string
		wantKey    string
		wantErr    bool
	}{
		{name: "nil request", nilRequest: true, wantKey: ""},
		{name: "ipv4 with port", remoteAddr: "203.0.113.5:1234", wantKey: "203.0.113.5"},
		{name: "ipv6 with port", remoteAddr: "[2001:db8::1]:443", wantKey: "2001:db8::1"},
		{name: "missing port", remoteAddr: "203.0.113.5", wantErr: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var request *http.Request
			if !testCase.nilRequest {
				request = &http.Request{RemoteAddr: testCase.remoteAddr}
			}

			key, err := DefaultGetRateLimitingKey(request)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if key != testCase.wantKey {
				t.Fatalf("got %q, want %q", key, testCase.wantKey)
			}
		})
	}
}
