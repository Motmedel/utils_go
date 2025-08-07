package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrEmptyPrivateKey = errors.New("empty private key")
	ErrEmptyPublicKey  = errors.New("empty public key")
	ErrSignatureMismatch = errors.New("signature mismatch")
)
