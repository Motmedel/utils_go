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
	ErrUnsupportedFileExtension              = errors.New("an unsupported file extension was encountered")
	ErrUnexpectedContentEncoding             = errors.New("an unexpected content encoding was encountered")
	ErrBadIfModifiedSinceTimestamp           = errors.New("bad If-Modified-Since timestamp")
)

type UnsupportedFileExtensionError struct {
	motmedelErrors.InputError
}

func (unsupportedFileExtensionError *UnsupportedFileExtensionError) Is(target error) bool {
	return target == ErrUnsupportedFileExtension
}

type UnexpectedContentEncodingError struct {
	motmedelErrors.InputError
}

func (unexpectedContentEncodingError *UnexpectedContentEncodingError) Is(target error) bool {
	return target == ErrUnexpectedContentEncoding
}

type BadIfModifiedSinceTimestamp struct {
	motmedelErrors.InputError
}

func (badIfModifiedSinceTimestamp *BadIfModifiedSinceTimestamp) Is(target error) bool {
	return target == ErrBadIfModifiedSinceTimestamp
}
