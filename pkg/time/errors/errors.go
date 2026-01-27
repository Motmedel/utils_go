package errors

import "errors"

var (
	ErrExpired          = errors.New("expired")
	ErrNegativeDuration = errors.New("negative duration")
)
