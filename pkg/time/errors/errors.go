package errors

import "errors"

var (
	ErrEmptyDuration = errors.New("empty duration")
	ErrExpired       = errors.New("expired")
)
