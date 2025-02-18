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
	ErrNilHostToMuxSpecification             = errors.New("nil host to mux specification")
	ErrNilMuxSpecification                   = errors.New("nil mux specification")
	ErrUnusableMuxSpecification              = errors.New("unusable mux specification")
)

type BadIfModifiedSinceTimestamp struct {
	motmedelErrors.Error
}

func (badIfModifiedSinceTimestamp *BadIfModifiedSinceTimestamp) Is(target error) bool {
	return target == ErrBadIfModifiedSinceTimestamp
}
