package hmac

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"hash"
)

type Method struct {
	Secret   []byte
	HashFunc func() hash.Hash
	Name     string
}

func (method *Method) Sign(message []byte) ([]byte, error) {
	secret := method.Secret
	if len(secret) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptySecret)
	}

	mac := hmac.New(method.HashFunc, secret)
	_, err := mac.Write(message)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(err)
	}

	return mac.Sum(nil), nil
}

func (method *Method) Verify(message []byte, signature []byte) error {
	secret := method.Secret
	if len(secret) == 0 {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptyPublicKey)
	}

	expectedMac := hmac.New(method.HashFunc, secret)
	_, err := expectedMac.Write(message)
	if err != nil {
		return motmedelErrors.NewWithTrace(err)
	}

	if hmac.Equal(expectedMac.Sum(nil), signature) {
		return nil
	} else {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
	}
}

func (method *Method) GetName() string {
	return method.Name
}

func New(algorithm string, secret []byte) (*Method, error) {
	switch algorithm {
	case "HS256":
		return &Method{Secret: secret, HashFunc: sha256.New, Name: "HS256"}, nil
	case "HS384":
		return &Method{Secret: secret, HashFunc: sha512.New384, Name: "HS384"}, nil
	case "HS512":
		return &Method{Secret: secret, HashFunc: sha512.New, Name: "HS512"}, nil
	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %q", motmedelCryptoErrors.ErrUnsupportedAlgorithm, algorithm),
		)
	}
}
