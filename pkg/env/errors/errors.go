package errors

import "errors"

var (
	ErrNotPresent = errors.New("non-present environment variable")
)
