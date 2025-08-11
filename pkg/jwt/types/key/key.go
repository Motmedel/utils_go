package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	motmedelMaps "github.com/Motmedel/utils_go/pkg/maps"
	"github.com/Motmedel/utils_go/pkg/utils"
	"math/big"
)

type Key struct {
	Kty string `json:"kty"`
}

type RsaKey struct {
	Key
	N string `json:"n"`
	E string `json:"e"`
}

func (k *RsaKey) PublicKey() (*rsa.PublicKey, error) {
	n := k.N
	nBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
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
		return nil, motmedelErrors.NewWithTrace(
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

	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: exponent}, nil
}

type EcKey struct {
	Key
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func (k *EcKey) PublicKey() (*ecdsa.PublicKey, error) {
	x := k.X
	xBytes, err := base64.RawURLEncoding.DecodeString(x)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf(
				"base64 raw url encoding decode string (x): %w",
				err,
			),
			x,
		)
	}

	y := k.Y
	yBytes, err := base64.RawURLEncoding.DecodeString(y)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf(
				"base64 raw url encoding decode string (y): %w",
				err,
			),
			y,
		)
	}

	crv := k.Crv
	curve := CurveFromCrv(crv)
	if utils.IsNil(curve) {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrUnsupportedCrv, crv)
	}

	return &ecdsa.PublicKey{Curve: curve, X: new(big.Int).SetBytes(xBytes), Y: new(big.Int).SetBytes(yBytes)}, nil
}

func RsaKeyFromMap(m map[string]any) (*RsaKey, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := motmedelMaps.MapGetConvert[string](m, "kty")
	if err != nil {
		return nil, fmt.Errorf("map get convert (kty): %w", err)
	}

	if kty != "RSA" {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrKtyMismatch)
	}

	n, err := motmedelMaps.MapGetConvert[string](m, "n")
	if err != nil {
		return nil, fmt.Errorf("map get convert (n): %w", err)
	}

	e, err := motmedelMaps.MapGetConvert[string](m, "e")
	if err != nil {
		return nil, fmt.Errorf("map get convert (e): %w", err)
	}

	return &RsaKey{Key: Key{Kty: kty}, N: n, E: e}, nil

}

func EcKeyFromMap(m map[string]any) (*EcKey, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := motmedelMaps.MapGetConvert[string](m, "kty")
	if err != nil {
		return nil, fmt.Errorf("map get convert (kty): %w", err)
	}

	if kty != "EC" {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrKtyMismatch)
	}

	crv, err := motmedelMaps.MapGetConvert[string](m, "crv")
	if err != nil {
		return nil, fmt.Errorf("map get convert (crv): %w", err)
	}

	x, err := motmedelMaps.MapGetConvert[string](m, "x")
	if err != nil {
		return nil, fmt.Errorf("map get convert (x): %w", err)
	}

	y, err := motmedelMaps.MapGetConvert[string](m, "y")
	if err != nil {
		return nil, fmt.Errorf("map get convert (y): %w", err)
	}

	return &EcKey{Key: Key{Kty: kty}, Crv: crv, X: x, Y: y}, nil
}

func CurveFromCrv(crv string) elliptic.Curve {
	switch crv {
	case "P-256":
		return elliptic.P256()
	case "P-384":
		return elliptic.P384()
	case "P-521":
		return elliptic.P521()
	}

	return nil
}
