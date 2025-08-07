package eddsa

import (
	"crypto/ed25519"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

const Name = "EdDSA"

type Method struct {
	PrivateKey []byte
	PublicKey  []byte
}

func (method *Method) Sign(message []byte) ([]byte, error) {
	privateKey := method.PrivateKey
	if len(privateKey) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptyPrivateKey)
	}

	return ed25519.Sign(privateKey, message), nil
}

func (method *Method) Verify(message []byte, signature []byte) error {
	publicKey := method.PublicKey
	if len(publicKey) == 0 {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptyPublicKey)
	}

	if ok := ed25519.Verify(publicKey, message, signature); ok {
		return nil
	} else {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
	}
}

func (method *Method) GetName() string {
	return Name
}
