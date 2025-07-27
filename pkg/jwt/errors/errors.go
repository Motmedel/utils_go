package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrNilToken         = errors.New("nil token")
	ErrEmptyTokenString = errors.New("empty token string")
	ErrEmptySigningKey  = errors.New("empty signing key")
	ErrInvalidToken     = errors.New("invalid token")
	ErrNilClaims        = errors.New("nil claims")
)
