package errors

import (
	"errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

var (
	ErrNilHandler                            = errors.New("handler is nil")
	ErrNoResponseWritten                     = errors.New("no response was written")
	ErrNilErrorResponseProblemDetail         = errors.New("the error response problem detail is nil")
	ErrUnsetErrorResponseProblemDetailStatus = errors.New("the error response problem detail status is unset")
	ErrBadIfModifiedSinceTimestamp           = errors.New("bad If-Modified-Since timestamp")
)

type BadIfModifiedSinceTimestamp struct {
	motmedelErrors.InputError
}

func (badIfModifiedSinceTimestamp *BadIfModifiedSinceTimestamp) Is(target error) bool {
	return target == ErrBadIfModifiedSinceTimestamp
}
