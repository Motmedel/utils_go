package client_side_encryption

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"testing"

	"github.com/Motmedel/utils_go/pkg/http/mux/utils/client_side_encryption/body_parser_config"
	"github.com/go-jose/go-jose/v4"
)

func makeJwe(t *testing.T, recipientPublicKey *ecdsa.PublicKey, keyId string) []byte {
	t.Helper()

	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.ECDH_ES, Key: recipientPublicKey, KeyID: keyId},
		nil,
	)
	if err != nil {
		t.Fatalf("jose new encrypter: %v", err)
	}

	jwe, err := encrypter.Encrypt([]byte(`{}`))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	serialized, err := jwe.CompactSerialize()
	if err != nil {
		t.Fatalf("compact serialize: %v", err)
	}

	return []byte(serialized)
}

func TestNewBodyParserSetsKeyIdentifier(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	bodyParser, err := NewBodyParser(privateKey, body_parser_config.WithKeyIdentifier("expected-kid"))
	if err != nil {
		t.Fatalf("new body parser: %v", err)
	}

	if bodyParser.KeyIdentifier != "expected-kid" {
		t.Errorf("key identifier = %q, want %q", bodyParser.KeyIdentifier, "expected-kid")
	}
}

func TestBodyParserParseKeyIdentifierMismatch(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	bodyParser, err := NewBodyParser(privateKey, body_parser_config.WithKeyIdentifier("expected-kid"))
	if err != nil {
		t.Fatalf("new body parser: %v", err)
	}

	_, responseError := bodyParser.Parse(nil, makeJwe(t, &privateKey.PublicKey, "other-kid"))
	if responseError == nil {
		t.Fatal("expected a response error")
	}

	problemDetail := responseError.ProblemDetail
	if problemDetail == nil {
		t.Fatal("expected a problem detail")
	}

	if problemDetail.Status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", problemDetail.Status, http.StatusBadRequest)
	}
}

func TestBodyParserParseUnusableKeyIsServerError(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	// An int is not a supported key type; `jwe.Decrypt` fails with
	// `jose.ErrUnsupportedKeyType` rather than `jose.ErrCryptoFailure`.
	bodyParser, err := NewBodyParser(12345)
	if err != nil {
		t.Fatalf("new body parser: %v", err)
	}

	_, responseError := bodyParser.Parse(nil, makeJwe(t, &privateKey.PublicKey, ""))
	if responseError == nil {
		t.Fatal("expected a response error")
	}

	if responseError.ServerError == nil {
		t.Error("expected a server error")
	}

	if responseError.ClientError != nil {
		t.Errorf("unexpected client error: %v", responseError.ClientError)
	}
}

func TestBodyParserParseDecryptFailureIsClientError(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	otherPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}

	bodyParser, err := NewBodyParser(privateKey)
	if err != nil {
		t.Fatalf("new body parser: %v", err)
	}

	_, responseError := bodyParser.Parse(nil, makeJwe(t, &otherPrivateKey.PublicKey, ""))
	if responseError == nil {
		t.Fatal("expected a response error")
	}

	if responseError.ServerError != nil {
		t.Errorf("unexpected server error: %v", responseError.ServerError)
	}

	if responseError.ClientError == nil {
		t.Error("expected a client error")
	}

	problemDetail := responseError.ProblemDetail
	if problemDetail == nil {
		t.Fatal("expected a problem detail")
	}

	if problemDetail.Status != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", problemDetail.Status, http.StatusBadRequest)
	}
}
