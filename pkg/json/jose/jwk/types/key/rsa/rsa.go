package rsa

import (
	"crypto"
	rsa2 "crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	"github.com/Motmedel/utils_go/pkg/maps"
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
	for i := 0; i < len(eBytes); i++ {
		exponent = exponent<<8 + int(eBytes[i])
	}

	return &rsa2.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}, nil
}

func New(m map[string]any) (*Key, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := maps.MapGetConvert[string](m, "kty")
	if err != nil {
		return nil, fmt.Errorf("map get convert (kty): %w", err)
	}

	if kty != "RSA" {
		return nil, errors.NewWithTrace(motmedelJwkErrors.ErrKtyMismatch)
	}

	n, err := maps.MapGetConvert[string](m, "n")
	if err != nil {
		return nil, fmt.Errorf("map get convert (n): %w", err)
	}

	e, err := maps.MapGetConvert[string](m, "e")
	if err != nil {
		return nil, fmt.Errorf("map get convert (e): %w", err)
	}

	return &Key{N: n, E: e}, nil

}
