package errors

import "errors"

var (
	ErrKtyMismatch    = errors.New("kty mismatch")
	ErrUnsupportedCrv = errors.New("unsupported crv")
	ErrUnsupportedKty = errors.New("unsupported kty")
	ErrUnknownKeyId   = errors.New("unknown key id")
)
