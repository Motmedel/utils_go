package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrEmptyPrivateKey      = errors.New("empty private key")
	ErrEmptyPublicKey       = errors.New("empty public key")
	ErrEmptySecret          = errors.New("empty secret")
	ErrSignatureMismatch    = errors.New("signature mismatch")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrNilSigner            = errors.New("nil signer")
	ErrNilVerifier          = errors.New("nil verifier")
	ErrNilMethod            = errors.New("nil method")
	ErrCurveMismatch        = errors.New("curve mismatch")
	ErrUnsupportedCurve     = errors.New("unsupported curve")
	ErrNilCurve             = errors.New("nil curve")
	ErrNilCurveParams       = errors.New("nil curve params")
)
