package ec

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/Motmedel/utils_go/pkg/errors"
	errors2 "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/types/key"
	"github.com/Motmedel/utils_go/pkg/maps"
	"github.com/Motmedel/utils_go/pkg/utils"
)

func curveFromCrv(crv string) elliptic.Curve {
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

type Key struct {
	key.Key
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func (k *Key) PublicKey() (*ecdsa.PublicKey, error) {
	x := k.X
	xBytes, err := base64.RawURLEncoding.DecodeString(x)
	if err != nil {
		return nil, errors.NewWithTrace(
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
		return nil, errors.NewWithTrace(
			fmt.Errorf(
				"base64 raw url encoding decode string (y): %w",
				err,
			),
			y,
		)
	}

	crv := k.Crv
	curve := curveFromCrv(crv)
	if utils.IsNil(curve) {
		return nil, errors.NewWithTrace(errors2.ErrUnsupportedCrv, crv)
	}

	return &ecdsa.PublicKey{Curve: curve, X: new(big.Int).SetBytes(xBytes), Y: new(big.Int).SetBytes(yBytes)}, nil
}

func New(m map[string]any) (*Key, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := maps.MapGetConvert[string](m, "kty")
	if err != nil {
		return nil, fmt.Errorf("map get convert (kty): %w", err)
	}

	if kty != "EC" {
		return nil, errors.NewWithTrace(errors2.ErrKtyMismatch)
	}

	crv, err := maps.MapGetConvert[string](m, "crv")
	if err != nil {
		return nil, fmt.Errorf("map get convert (crv): %w", err)
	}

	x, err := maps.MapGetConvert[string](m, "x")
	if err != nil {
		return nil, fmt.Errorf("map get convert (x): %w", err)
	}

	y, err := maps.MapGetConvert[string](m, "y")
	if err != nil {
		return nil, fmt.Errorf("map get convert (y): %w", err)
	}

	return &Key{Key: key.Key{Kty: kty}, Crv: crv, X: x, Y: y}, nil
}
