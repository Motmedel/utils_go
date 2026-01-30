package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/json"
	"fmt"

	motmedelCryptoEcdsa "github.com/Motmedel/utils_go/pkg/crypto/ecdsa"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelCryptoRsa "github.com/Motmedel/utils_go/pkg/crypto/rsa"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	ecKey "github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key/ec"
	rsaKey "github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key/rsa"
	"github.com/Motmedel/utils_go/pkg/utils"
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

// MarshalJSON ensures that the material fields (e.g., EC or RSA parameters)
// are serialized at the same top-level as the common JWK fields (alg, kty, kid, use).
func (k *Key) MarshalJSON() ([]byte, error) {
	// Start with base fields
	m := make(map[string]any)
	if k.Alg != "" {
		m["alg"] = k.Alg
	}
	if k.Kty != "" {
		m["kty"] = k.Kty
	}
	if k.Kid != "" {
		m["kid"] = k.Kid
	}
	if k.Use != "" {
		m["use"] = k.Use
	}

	// Merge material (if present) at the same level
	if k.Material != nil {
		b, err := json.Marshal(k.Material)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("json marshal (material): %w", err), k.Material)
		}
		var mat map[string]any
		if err := json.Unmarshal(b, &mat); err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("json unmarshal (material to map): %w", err), string(b))
		}
		for key, val := range mat {
			m[key] = val
		}
	}

	return json.Marshal(m)
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

func New(m map[string]any) (*Key, error) {
	if m == nil {
		return nil, nil
	}

	kty, err := utils.MapGetConvert[string](m, "kty")
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
