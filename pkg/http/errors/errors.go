package errors

import (
	"errors"
	"strconv"
)

var (
	ErrNilHttpClient               = errors.New("nil client")
	ErrNilHttpRequest              = errors.New("nil request")
	ErrNilHttpResponse             = errors.New("nil response")
	ErrNilHttpRequestHeader        = errors.New("nil request header")
	ErrNilHttpRequestUrl           = errors.New("nil request url")
	ErrNilHttpResponseHeader       = errors.New("nil response header")
	ErrNilHttpRequestBodyReader    = errors.New("nil http request body reader")
	ErrNilHttpResponseBodyReader   = errors.New("nil http response body reader")
	ErrNilHttpContext              = errors.New("nil http context")
	ErrNon2xxStatusCode            = errors.New("non-2xx status code")
	ErrEmptyMethod                 = errors.New("empty method")
	ErrEmptyUrl                    = errors.New("empty url")
	ErrEmptyResponseBody           = errors.New("empty response body")
	ErrReattemptFailedError        = errors.New("reattempt failed")
	ErrBadIfModifiedSinceTimestamp = errors.New("bad If-Modified-Since timestamp")
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

func (non2xxStatusCodeError *Non2xxStatusCodeError) GetCode() string {
	if non2xxStatusCodeError.StatusCode == 0 {
		return ""
	}
	return strconv.Itoa(non2xxStatusCodeError.StatusCode)
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
