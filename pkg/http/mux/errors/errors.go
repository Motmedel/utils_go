package errors

import (
	"errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

var (
	ErrNilHandler                            = errors.New("nil handler")
	ErrNoResponseWritten                     = errors.New("no response was written")
	ErrNilErrorResponseProblemDetail         = errors.New("nil error response problem detail")
	ErrUnsetErrorResponseProblemDetailStatus = errors.New("the error response problem detail status is unset")
	ErrBadIfModifiedSinceTimestamp           = errors.New("bad If-Modified-Since timestamp")
	ErrNoResponseWriterFlusher               = errors.New("no response writer flusher")
	ErrTransferEncodingAlreadySet            = errors.New("transfer encoding already set")
)

type BadIfModifiedSinceTimestamp struct {
	motmedelErrors.InputError
}

func (badIfModifiedSinceTimestamp *BadIfModifiedSinceTimestamp) Is(target error) bool {
	return target == ErrBadIfModifiedSinceTimestamp
}
