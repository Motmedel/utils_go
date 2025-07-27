package errors

import "errors"

var (
	ErrEmptyAddress    = errors.New("empty address")
	ErrAddressMismatch = errors.New("address mismatch")
)
