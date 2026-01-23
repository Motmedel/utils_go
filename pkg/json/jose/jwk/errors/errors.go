package errors

import "errors"

var (
	ErrKtyMismatch    = errors.New("kty mismatch")
	ErrUnsupportedCrv = errors.New("unsupported crv")
	ErrUnsupportedKty = errors.New("unsupported kty")
	ErrNilKey         = errors.New("nil key")
	ErrEmptyAlg       = errors.New("empty alg")
)
