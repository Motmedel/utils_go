package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"

	motmedelCryptoEcdsa "github.com/Motmedel/utils_go/pkg/crypto/ecdsa"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelCryptoRsa "github.com/Motmedel/utils_go/pkg/crypto/rsa"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	ecKey "github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key/ec"
	rsaKey "github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key/rsa"
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

func (k *Key) NamedVerifier() (motmedelCryptoInterfaces.NamedVerifier, error) {
	material := k.Material
	if material == nil {
		return nil, nil
	}

	publicKey, err := material.PublicKey()
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("public key: %w", err), material)
	}

	switch typedPublicKey := publicKey.(type) {
	case *ecdsa.PublicKey:
		method, err := motmedelCryptoEcdsa.FromPublicKey(typedPublicKey)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("ecdsa from public key: %w", err), typedPublicKey)
		}
		return method, nil
	case *rsa.PublicKey:
		alg := k.Alg
		if alg == "" {
			return nil, motmedelErrors.NewWithTrace(motmedelJwkErrors.ErrEmptyAlg)
		}

		method, err := motmedelCryptoRsa.New(alg, nil, typedPublicKey)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("rsa new: %w", err), alg)
		}
		return method, nil
	default:
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("unsupported public key type: %T", publicKey))
	}
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
		material, err = rsaKey.New(m)
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("rsa new: %w", err), m)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return nil, wrappedErr
		}
	case "EC":
		material, err = ecKey.New(m)
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
