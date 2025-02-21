package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNilEndpointSpecification    = errors.New("nil endpoint specification")
	ErrNoResponseWritten           = errors.New("no response was written")
	ErrBadIfModifiedSinceTimestamp = errors.New("bad If-Modified-Since timestamp")
	ErrNoResponseWriterFlusher     = errors.New("no response writer flusher")
	ErrTransferEncodingAlreadySet  = errors.New("transfer encoding already set")
	ErrNilHostToMuxSpecification   = errors.New("nil host to mux specification")
	ErrNilMuxSpecification         = errors.New("nil mux specification")
	ErrUnusableMuxSpecification    = errors.New("unusable mux specification")
	ErrNilResponseWriter           = errors.New("nil response writer")
	ErrCouldNotObtainHttpContext   = errors.New("could not obtain http context")
	ErrNilStaticContent            = errors.New("nil static content")
	ErrNilContentEncodingToData    = errors.New("nil content-encoding to data")
	// TODO: Move to problem detail errors
	ErrNilProblemDetail            = errors.New("nil problem detail")
	ErrEmptyStatus                 = errors.New("empty status")
	ErrUnusableResponseError       = errors.New("unusable response error")
	ErrEmptyResponseErrorErrors    = errors.New("empty response error errors")
	ErrMultipleResponseErrorErrors = fmt.Errorf("%w: multiple response error errors", ErrUnusableMuxSpecification)
	ErrUnexpectedResponseErrorType = fmt.Errorf("%w: unexpected response error type", ErrUnusableResponseError)
	ErrNilMux                      = errors.New("nil mux")
	ErrNilVhostMux                 = errors.New("nil vhost mux")
)
