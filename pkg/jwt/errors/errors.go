package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrNilToken         = errors.New("nil token")
	ErrEmptyTokenString = errors.New("empty token string")
	ErrEmptySigningKey  = errors.New("empty signing key")
	ErrNilMethod = errors.New("nil method")
	ErrEmptyParameterName = errors.New("empty parameter name")
	ErrNilValidator = errors.New("nil validator")
)
