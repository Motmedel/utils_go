package errors

import (
	"errors"
	"fmt"
)

var (
	ErrNilEndpointSpecification      = errors.New("nil endpoint specification")
	ErrNoResponseWritten             = errors.New("no response was written")
	ErrNoResponseWriterFlusher       = errors.New("no response writer flusher")
	ErrTransferEncodingAlreadySet    = errors.New("transfer encoding already set")
	ErrNilHostToMuxSpecification     = errors.New("nil host to mux specification")
	ErrNilMuxSpecification           = errors.New("nil mux specification")
	ErrUnusableMuxSpecification      = errors.New("unusable mux specification")
	ErrNilResponseWriter             = errors.New("nil response writer")
	ErrCouldNotObtainHttpContext     = errors.New("could not obtain http context")
	ErrNilStaticContent              = errors.New("nil static content")
	ErrNilContentEncodingToData      = errors.New("nil content-encoding to data")
	ErrContentEncodingToDataNotOk    = errors.New("content encoding to data not ok")
	ErrNilStaticContentData          = errors.New("nil static content data")
	ErrEmptyStatus                   = errors.New("empty status")
	ErrUnusableResponseError         = errors.New("unusable response error")
	ErrEmptyResponseErrorErrors      = errors.New("empty response error errors")
	ErrMultipleResponseErrorErrors   = fmt.Errorf("%w: multiple response error errors", ErrUnusableMuxSpecification)
	ErrUnexpectedResponseErrorType   = fmt.Errorf("%w: unexpected response error type", ErrUnusableResponseError)
	ErrNilMux                        = errors.New("nil mux")
	ErrNilVhostMux                   = errors.New("nil vhost mux")
	ErrEmptyResponseErrorContentType = errors.New("empty response error content type")
	ErrUnexpectedContentEncoding     = errors.New("unexpected content encoding")
	ErrUnsupportedFileExtension      = errors.New("unsupported file extension")
	ErrNilHeaderParameter            = errors.New("nil header parameter")
	ErrEmptyContentType              = errors.New("empty content type")
	ErrNilRequestParser              = errors.New("nil request parser")
)
