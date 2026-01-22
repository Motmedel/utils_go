package key

import (
	"crypto"
	"fmt"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key/ec"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key/rsa"
	motmedelMaps "github.com/Motmedel/utils_go/pkg/maps"
)

type Key struct {
	Alg string `json:"alg,omitempty"`
	Kty string `json:"kty,omitempty"`
	Kid string `json:"kid,omitempty"`
	Use string `json:"use,omitempty"`

	Material interface {
		PublicKey() (crypto.PublicKey, error)
	} `json:"-"`
}

type Keys struct {
	Keys []map[string]any `json:"keys,omitempty"`
}

func New(m map[string]any) (*Key, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := motmedelMaps.MapGetConvert[string](m, "kty")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert: %w", err), m)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
		return nil, wrappedErr
	}

	var material interface {
		PublicKey() (crypto.PublicKey, error)
	}

	switch kty {
	case "RSA":
		material, err = rsa.New(m)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("rsa new: %w", err), m)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return nil, wrappedErr
		}
	case "EC":
		material, err = ec.New(m)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("ec new: %w", err), m)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return nil, wrappedErr
		}
	default:
		return nil, motmedelErrors.NewWithTrace(motmedelJwkErrors.ErrUnsupportedKty, kty)
	}

	alg, _ := m["alg"].(string)
	kid, _ := m["kid"].(string)
	use, _ := m["use"].(string)

	return &Key{Alg: alg, Kty: kty, Kid: kid, Use: use, Material: material}, nil
}
