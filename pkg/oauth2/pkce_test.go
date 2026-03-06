package oauth2

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGenerateVerifier(t *testing.T) {
	verifier := GenerateVerifier()

	if len(verifier) != 43 {
		t.Errorf("expected verifier length 43, got %d", len(verifier))
	}

	// Verify uniqueness (two calls should produce different results).
	verifier2 := GenerateVerifier()
	if verifier == verifier2 {
		t.Errorf("expected unique verifiers, got identical values")
	}
}

func TestS256ChallengeFromVerifier(t *testing.T) {
	// Use a known verifier and expected challenge.
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	sha := sha256.Sum256([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(sha[:])

	challenge := S256ChallengeFromVerifier(verifier)
	if challenge != expected {
		t.Errorf("expected challenge %q, got %q", expected, challenge)
	}
}

func TestS256ChallengeOption(t *testing.T) {
	verifier := "test_verifier"
	opts := S256ChallengeOption(verifier)

	if len(opts) != 2 {
		t.Fatalf("expected 2 options, got %d", len(opts))
	}

	if opts[0].key != codeChallengeMethodKey || opts[0].value != "S256" {
		t.Errorf("expected challenge method option, got key=%q value=%q", opts[0].key, opts[0].value)
	}

	expectedChallenge := S256ChallengeFromVerifier(verifier)
	if opts[1].key != codeChallengeKey || opts[1].value != expectedChallenge {
		t.Errorf("expected challenge option, got key=%q value=%q", opts[1].key, opts[1].value)
	}
}

func TestVerifierOption(t *testing.T) {
	verifier := "test_verifier"
	opt := VerifierOption(verifier)

	if opt.key != codeVerifierKey {
		t.Errorf("expected key %q, got %q", codeVerifierKey, opt.key)
	}
	if opt.value != verifier {
		t.Errorf("expected value %q, got %q", verifier, opt.value)
	}
}
