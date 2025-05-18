package errors

import "errors"

var (
	ErrNotPresent = errors.New("non-present environment variable")
	ErrEmpty      = errors.New("empty environment variable")
)
