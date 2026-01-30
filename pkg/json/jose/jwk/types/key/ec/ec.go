package ec

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
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
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

func (k *Key) PublicKey() (crypto.PublicKey, error) {
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
		return nil, errors.NewWithTrace(motmedelJwkErrors.ErrUnsupportedCrv, crv)
	}

	return &ecdsa.PublicKey{Curve: curve, X: new(big.Int).SetBytes(xBytes), Y: new(big.Int).SetBytes(yBytes)}, nil
}

func New(m map[string]any) (*Key, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := utils.MapGetConvert[string](m, "kty")
	if err != nil {
		return nil, fmt.Errorf("map get convert (kty): %w", err)
	}

	if kty != "EC" {
		return nil, errors.NewWithTrace(motmedelJwkErrors.ErrKtyMismatch)
	}

	crv, err := utils.MapGetConvert[string](m, "crv")
	if err != nil {
		return nil, fmt.Errorf("map get convert (crv): %w", err)
	}

	x, err := utils.MapGetConvert[string](m, "x")
	if err != nil {
		return nil, fmt.Errorf("map get convert (x): %w", err)
	}

	y, err := utils.MapGetConvert[string](m, "y")
	if err != nil {
		return nil, fmt.Errorf("map get convert (y): %w", err)
	}

	return &Key{Crv: crv, X: x, Y: y}, nil
}

func padLeft(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}

func NewFromPublicKey(publicKey *ecdsa.PublicKey) (*Key, error) {
	if publicKey == nil {
		return nil, nil
	}

	curve := publicKey.Curve
	if utils.IsNil(curve) {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("public key curve"))
	}

	x := publicKey.X
	if x == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("public key x"))
	}

	y := publicKey.Y
	if y == nil {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("public key y"))
	}

	var crv string
	var size int
	switch publicKey.Curve {
	case elliptic.P256():
		crv = "P-256"
		size = 32
	case elliptic.P384():
		crv = "P-384"
		size = 48
	case elliptic.P521():
		crv = "P-521"
		size = 66
	default:
		return nil, fmt.Errorf("unsupported curve: %T", publicKey.Curve)
	}

	return &Key{
		Crv: crv,
		X:   base64.RawURLEncoding.EncodeToString(padLeft(x.Bytes(), size)),
		Y:   base64.RawURLEncoding.EncodeToString(padLeft(y.Bytes(), size)),
	}, nil
}
