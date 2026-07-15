package cose

import (
	"bytes"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type vectorKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	D   string `json:"d"`
}

type vector struct {
	Title string `json:"title"`
	Input struct {
		Plaintext string `json:"plaintext"`
		Enveloped struct {
			Protected struct {
				Alg string `json:"alg"`
			} `json:"protected"`
			Recipients []struct {
				Key vectorKey `json:"key"`
			} `json:"recipients"`
		} `json:"enveloped"`
	} `json:"input"`
	Output struct {
		Cbor string `json:"cbor"`
	} `json:"output"`
}

func readVector(t *testing.T, name string) *vector {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read vector file: %v", err)
	}

	var parsedVector vector
	if err := json.Unmarshal(data, &parsedVector); err != nil {
		t.Fatalf("unmarshal vector: %v", err)
	}

	return &parsedVector
}

func vectorPrivateKey(t *testing.T, key *vectorKey) *ecdh.PrivateKey {
	t.Helper()

	if key.Crv != "P-256" {
		t.Fatalf("unexpected curve: %s", key.Crv)
	}

	d, err := base64.RawURLEncoding.DecodeString(key.D)
	if err != nil {
		t.Fatalf("decode d: %v", err)
	}

	privateKey, err := ecdh.P256().NewPrivateKey(d)
	if err != nil {
		t.Fatalf("new private key: %v", err)
	}

	return privateKey
}

func TestDecryptVectors(t *testing.T) {
	for _, vectorName := range []string{"p256-hkdf-256-01.json", "p256-hkdf-256-02.json"} {
		t.Run(vectorName, func(t *testing.T) {
			parsedVector := readVector(t, vectorName)

			message, err := hex.DecodeString(parsedVector.Output.Cbor)
			if err != nil {
				t.Fatalf("decode message hex: %v", err)
			}

			key := parsedVector.Input.Enveloped.Recipients[0].Key
			privateKey := vectorPrivateKey(t, &key)

			result, err := Decrypt(message, privateKey, nil)
			if err != nil {
				t.Fatalf("decrypt: %v", err)
			}

			if expectedPlaintext := parsedVector.Input.Plaintext; string(result.Plaintext) != expectedPlaintext {
				t.Errorf("plaintext: got %q, want %q", result.Plaintext, expectedPlaintext)
			}

			if expectedKid := key.Kid; string(result.KeyIdentifier) != expectedKid {
				t.Errorf("key identifier: got %q, want %q", result.KeyIdentifier, expectedKid)
			}
		})
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	for _, algorithm := range []Algorithm{AlgorithmA128GCM, AlgorithmA192GCM, AlgorithmA256GCM} {
		privateKey, err := ecdh.P256().GenerateKey(nil)
		if err != nil {
			t.Fatalf("generate key: %v", err)
		}

		plaintext := []byte("This is the content.")
		keyIdentifier := []byte("test-key-id")

		message, err := Encrypt(
			plaintext,
			privateKey.PublicKey(),
			&EncryptOptions{
				ContentEncryptionAlgorithm: algorithm,
				KeyIdentifier:              keyIdentifier,
				ContentType:                "application/cbor",
			},
		)
		if err != nil {
			t.Fatalf("encrypt: %v", err)
		}

		result, err := Decrypt(message, privateKey, nil)
		if err != nil {
			t.Fatalf("decrypt: %v", err)
		}

		if !bytes.Equal(result.Plaintext, plaintext) {
			t.Errorf("plaintext: got %q, want %q", result.Plaintext, plaintext)
		}

		if !bytes.Equal(result.KeyIdentifier, keyIdentifier) {
			t.Errorf("key identifier: got %q, want %q", result.KeyIdentifier, keyIdentifier)
		}

		if contentType, ok := result.ContentType.(string); !ok || contentType != "application/cbor" {
			t.Errorf("content type: got %v, want application/cbor", result.ContentType)
		}
	}
}

func TestEncryptDecryptExternalAad(t *testing.T) {
	privateKey, err := ecdh.P256().GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	plaintext := []byte("This is the content.")
	externalAad := []byte("external")

	message, err := Encrypt(plaintext, privateKey.PublicKey(), &EncryptOptions{ExternalAad: externalAad})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	result, err := Decrypt(message, privateKey, &DecryptOptions{ExternalAad: externalAad})
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(result.Plaintext, plaintext) {
		t.Errorf("plaintext: got %q, want %q", result.Plaintext, plaintext)
	}

	if _, err := Decrypt(message, privateKey, nil); !errors.Is(err, ErrNoUsableRecipient) {
		t.Errorf("expected no usable recipient error with missing external aad, got %v", err)
	}
}

// TestDecryptTypescriptFixture decrypts a message produced by the @altshiftab/utils/cose
// TypeScript implementation, using the recipient key from vector p256-hkdf-256-01.
func TestDecryptTypescriptFixture(t *testing.T) {
	hexData, err := os.ReadFile(filepath.Join("testdata", "typescript_encrypted.hex"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	message, err := hex.DecodeString(strings.TrimSpace(string(hexData)))
	if err != nil {
		t.Fatalf("decode fixture hex: %v", err)
	}

	d, err := base64.RawURLEncoding.DecodeString("r_kHyZ-a06rmxM3yESK84r1otSg-aQcVStkRhA-iCM8")
	if err != nil {
		t.Fatalf("decode d: %v", err)
	}

	privateKey, err := ecdh.P256().NewPrivateKey(d)
	if err != nil {
		t.Fatalf("new private key: %v", err)
	}

	result, err := Decrypt(message, privateKey, nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	if expectedPlaintext := "Interop test content."; string(result.Plaintext) != expectedPlaintext {
		t.Errorf("plaintext: got %q, want %q", result.Plaintext, expectedPlaintext)
	}

	if expectedKid := "interop-key-id"; string(result.KeyIdentifier) != expectedKid {
		t.Errorf("key identifier: got %q, want %q", result.KeyIdentifier, expectedKid)
	}

	if contentType, ok := result.ContentType.(string); !ok || contentType != "application/cbor" {
		t.Errorf("content type: got %v, want application/cbor", result.ContentType)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	privateKey, err := ecdh.P256().GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	otherPrivateKey, err := ecdh.P256().GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	message, err := Encrypt([]byte("This is the content."), privateKey.PublicKey(), nil)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := Decrypt(message, otherPrivateKey, nil); !errors.Is(err, ErrNoUsableRecipient) {
		t.Errorf("expected no usable recipient error with wrong key, got %v", err)
	}
}
