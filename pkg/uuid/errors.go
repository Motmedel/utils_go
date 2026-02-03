package uuid

import "errors"

var (
	ErrInvalidUUIDLength = errors.New("invalid UUID length")
	ErrInvalidUUIDFormat = errors.New("invalid UUID format")
)
