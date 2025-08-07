package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrEmptyPrivateKey = errors.New("empty private key")
	ErrEmptyPublicKey  = errors.New("empty public key")
	ErrSignatureMismatch = errors.New("signature mismatch")
	ErrNilSigner = errors.New("nil signer")
	ErrNilVerifier = errors.New("nil verifier")
	ErrNilMethod = errors.New("nil method")
)
