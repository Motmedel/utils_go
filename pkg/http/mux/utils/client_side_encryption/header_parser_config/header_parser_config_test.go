package header_parser_config

import (
	"testing"

	"github.com/go-jose/go-jose/v4"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
	if defaults.HeaderName != DefaultHeaderName {
		t.Errorf("default header name = %q, want %q", defaults.HeaderName, DefaultHeaderName)
	}
	if defaults.KeyAlgorithm != DefaultKeyAlgorithm {
		t.Errorf("default key algorithm = %q, want %q", defaults.KeyAlgorithm, DefaultKeyAlgorithm)
	}
	if defaults.ContentEncryption != DefaultContentEncryption {
		t.Errorf("default content encryption = %q, want %q", defaults.ContentEncryption, DefaultContentEncryption)
	}
	if defaults.EncrypterOptions != DefaultEncrypterOptions {
		t.Errorf("default encrypter options = %p, want %p", defaults.EncrypterOptions, DefaultEncrypterOptions)
	}

	encrypterOptions := (&jose.EncrypterOptions{}).WithContentType("application/octet-stream")
	config := New(
		WithHeaderName("X-Custom-Jwk"),
		WithKeyAlgorithm(jose.ECDH_ES_A256KW),
		WithContentEncryption(jose.A128GCM),
		WithEncrypterOptions(encrypterOptions),
	)
	if config.HeaderName != "X-Custom-Jwk" {
		t.Errorf("header name = %q, want %q", config.HeaderName, "X-Custom-Jwk")
	}
	if config.KeyAlgorithm != jose.ECDH_ES_A256KW {
		t.Errorf("key algorithm = %q, want %q", config.KeyAlgorithm, jose.ECDH_ES_A256KW)
	}
	if config.ContentEncryption != jose.A128GCM {
		t.Errorf("content encryption = %q, want %q", config.ContentEncryption, jose.A128GCM)
	}
	if config.EncrypterOptions != encrypterOptions {
		t.Errorf("encrypter options = %p, want %p", config.EncrypterOptions, encrypterOptions)
	}
}
