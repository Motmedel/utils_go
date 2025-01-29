package errors

import (
	"errors"
)

var (
	ErrNilHttpClient         = errors.New("nil http client")
	ErrNilHttpRequest        = errors.New("nil http request")
	ErrNilHttpResponse       = errors.New("nil http response")
	ErrNilHttpRequestHeader  = errors.New("nil http request header")
	ErrNilHttpResponseHeader = errors.New("nil http response header")
	ErrNon2xxStatusCode      = errors.New("non-2xx status code")
	ErrEmptyMethod           = errors.New("empty http method")
	ErrEmptyUrl              = errors.New("empty url")
	ErrReattemptFailedError  = errors.New("reattempt failed")
)

type Non2xxStatusCodeError struct {
	StatusCode int
}

func (non2xxStatusCodeError *Non2xxStatusCodeError) Is(target error) bool {
	return target == ErrNon2xxStatusCode
}

func (non2xxStatusCodeError *Non2xxStatusCodeError) Error() string {
	return ErrNon2xxStatusCode.Error()
}

func (non2xxStatusCodeError *Non2xxStatusCodeError) GetInput() any {
	return non2xxStatusCodeError.StatusCode
}

type ReattemptFailedError struct {
	Attempt int
	Cause   error
}

func (reattemptFailedError *ReattemptFailedError) Is(target error) bool {
	return target == ErrReattemptFailedError
}

func (reattemptFailedError *ReattemptFailedError) Error() string {
	return ErrReattemptFailedError.Error()
}

func (reattemptFailedError *ReattemptFailedError) GetCause() error {
	return reattemptFailedError.Cause
}

func (reattemptFailedError *ReattemptFailedError) Unwrap() error {
	return reattemptFailedError.Cause
}
