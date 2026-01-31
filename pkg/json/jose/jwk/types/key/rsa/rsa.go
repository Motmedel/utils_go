package rsa

import (
	"crypto"
	rsa2 "crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type Key struct {
	N string `json:"n"`
	E string `json:"e"`
}

func (k *Key) PublicKey() (crypto.PublicKey, error) {
	n := k.N
	nBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, errors.NewWithTrace(
			fmt.Errorf(
				"base64 raw url encoding decode string (n): %w",
				err,
			),
			n,
		)
	}

	e := k.E
	eBytes, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, errors.NewWithTrace(
			fmt.Errorf(
				"base64 raw url encoding decode string (e): %w",
				err,
			),
			e,
		)
	}

	var exponent int
	for i := range eBytes {
		exponent = exponent<<8 + int(eBytes[i])
	}

	return &rsa2.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}, nil
}

func New(m map[string]any) (*Key, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := utils.MapGetConvert[string](m, "kty")
	if err != nil {
		return nil, fmt.Errorf("map get convert (kty): %w", err)
	}

	if kty != "RSA" {
		return nil, errors.NewWithTrace(motmedelJwkErrors.ErrKtyMismatch)
	}

	n, err := utils.MapGetConvert[string](m, "n")
	if err != nil {
		return nil, fmt.Errorf("map get convert (n): %w", err)
	}

	e, err := utils.MapGetConvert[string](m, "e")
	if err != nil {
		return nil, fmt.Errorf("map get convert (e): %w", err)
	}

	return &Key{N: n, E: e}, nil
}

// intToBigEndianBytes encodes an int into a minimal-length big-endian byte slice.
func intToBigEndianBytes(e int) []byte {
	if e == 0 {
		return []byte{0}
	}
	tmp := e
	n := 0
	for tmp > 0 {
		n++
		tmp >>= 8
	}
	b := make([]byte, n)
	for i := n - 1; i >= 0; i-- {
		b[i] = byte(e & 0xff)
		e >>= 8
	}
	return b
}

// NewFromPublicKey constructs RSA JWK material from a Go *rsa.PublicKey.
// It encodes N as base64url(big-endian bytes) and E as base64url(minimal big-endian bytes).
func NewFromPublicKey(publicKey *rsa2.PublicKey) (*Key, error) {
	if publicKey == nil {
		return nil, nil
	}

	if publicKey.N == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("public key N"))
	}

	nB64 := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	eB64 := base64.RawURLEncoding.EncodeToString(intToBigEndianBytes(publicKey.E))

	return &Key{N: nB64, E: eB64}, nil
}

// ThumbprintInput returns the RFC 7638 canonical JSON string used to compute
// the JWK Thumbprint for an RSA key: {"e":"%s","kty":"RSA","n":"%s"}
// The fields must be ordered exactly as above and values must be base64url-encoded
// without padding.
func (k *Key) ThumbprintInput() (string, error) {
	if k == nil {
		return "", nil
	}
	return fmt.Sprintf("{\"e\":\"%s\",\"kty\":\"RSA\",\"n\":\"%s\"}", k.E, k.N), nil
}
