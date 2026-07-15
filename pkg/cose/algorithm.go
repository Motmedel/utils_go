package cose

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

// Algorithm is a COSE algorithm identifier as registered in the IANA COSE Algorithms registry.
type Algorithm int64

const (
	AlgorithmA128GCM       Algorithm = 1
	AlgorithmA192GCM       Algorithm = 2
	AlgorithmA256GCM       Algorithm = 3
	AlgorithmEcdhEsHkdf256 Algorithm = -25
)

// ContentEncryption describes a COSE content-encryption algorithm.
type ContentEncryption struct {
	KeyBits int
	NewAead func(key []byte) (cipher.AEAD, error)
}

func newAesGcmAead(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes new cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher new gcm: %w", err)
	}

	return aead, nil
}

var contentEncryptionRegistry = map[Algorithm]*ContentEncryption{
	AlgorithmA128GCM: {KeyBits: 128, NewAead: newAesGcmAead},
	AlgorithmA192GCM: {KeyBits: 192, NewAead: newAesGcmAead},
	AlgorithmA256GCM: {KeyBits: 256, NewAead: newAesGcmAead},
}

// RegisterContentEncryption makes a content-encryption algorithm available to Encrypt and Decrypt.
func RegisterContentEncryption(algorithm Algorithm, contentEncryption *ContentEncryption) {
	contentEncryptionRegistry[algorithm] = contentEncryption
}
