package errors

import "errors"

var (
	ErrNilProblemDetail = errors.New("nil problem detail")
	ErrEmptyType        = errors.New("empty type")
	ErrEmptyTitle       = errors.New("empty title")
	ErrEmptyStatus      = errors.New("empty status")
	ErrNilEncoder       = errors.New("nil encoder")
)
