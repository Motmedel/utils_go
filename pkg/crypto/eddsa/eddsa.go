package eddsa

import (
	"crypto/ed25519"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

const Name = "EdDSA"

type Method struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
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

	if privateKey := method.PrivateKey; len(publicKey) == 0 && len(privateKey) != 0 {
		var err error
		privateKeyPublic := privateKey.Public()
		publicKey, err = utils.Convert[ed25519.PublicKey](privateKeyPublic)
		if err != nil {
			return motmedelErrors.NewWithTrace(
				fmt.Errorf("convert (private key public): %w", err),
				privateKeyPublic,
			)
		}
	}

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
