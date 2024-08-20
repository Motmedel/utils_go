package errors

import "errors"

var (
	ErrNilHandler                            = errors.New("handler is nil")
	ErrNoResponseWritten                     = errors.New("no response was written")
	ErrNilErrorResponseProblemDetail         = errors.New("the error response problem detail is nil")
	ErrUnsetErrorResponseProblemDetailStatus = errors.New("the error response problem detail status is unset")
)
