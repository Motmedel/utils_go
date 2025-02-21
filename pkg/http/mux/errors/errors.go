package errors

import (
	"errors"
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
	// TODO: Put in HTTP errors?
	ErrNilContentType = errors.New("nil content type")
	// TODO: Put in HTTP errors?
	ErrNilHttpRequestBodyReader  = errors.New("nil http request body reader")
	ErrCouldNotObtainHttpContext = errors.New("could not obtain http context")
	ErrNilStaticContent          = errors.New("nil static content")
	// TODO: Put in HTTP errors?
	ErrNilAcceptEncoding        = errors.New("nil accept encoding")
	ErrNilContentEncodingToData = errors.New("nil content-encoding to data")
	// TODO: Move to problem detail errors
	ErrEmptyStatus = errors.New("empty status")
	// TODO: Move to problem detail errors?
	ErrNilProblemDetail            = errors.New("nil problem detail")
	ErrInvalidResponseError        = errors.New("invalid response error")
	ErrUnexpectedResponseErrorType = errors.New("unexpected response error type")

	ErrNilMux      = errors.New("nil mux")
	ErrNilVhostMux = errors.New("nil vhost mux")
)
