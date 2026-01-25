package errors

import "errors"

var (
	ErrExpExpired = errors.New("exp expired")
	ErrNbfBefore  = errors.New("nbf before")
	ErrIatBefore  = errors.New("iat before")
)
