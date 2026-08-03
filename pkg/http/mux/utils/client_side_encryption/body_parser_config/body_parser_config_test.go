package body_parser_config

import (
	"testing"

	"github.com/go-jose/go-jose/v4"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
	if defaults.KeyIdentifier != "" {
		t.Errorf("default key identifier = %q, want %q", defaults.KeyIdentifier, "")
	}
	if defaults.KeyAlgorithm != DefaultKeyAlgorithm {
		t.Errorf("default key algorithm = %q, want %q", defaults.KeyAlgorithm, DefaultKeyAlgorithm)
	}
	if defaults.ContentEncryption != DefaultContentEncryption {
		t.Errorf("default content encryption = %q, want %q", defaults.ContentEncryption, DefaultContentEncryption)
	}

	config := New(
		WithKeyIdentifier("kid-1"),
		WithKeyAlgorithm(jose.ECDH_ES_A256KW),
		WithContentEncryption(jose.A128GCM),
	)
	if config.KeyIdentifier != "kid-1" {
		t.Errorf("key identifier = %q, want %q", config.KeyIdentifier, "kid-1")
	}
	if config.KeyAlgorithm != jose.ECDH_ES_A256KW {
		t.Errorf("key algorithm = %q, want %q", config.KeyAlgorithm, jose.ECDH_ES_A256KW)
	}
	if config.ContentEncryption != jose.A128GCM {
		t.Errorf("content encryption = %q, want %q", config.ContentEncryption, jose.A128GCM)
	}
}
