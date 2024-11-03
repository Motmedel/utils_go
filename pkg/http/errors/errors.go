package errors

import "errors"

var (
	ErrNilHttpClient    = errors.New("nil http client")
	ErrNilHttpRequest   = errors.New("nil http request")
	ErrNilHttpResponse  = errors.New("nil http response")
	ErrNon2xxStatusCode = errors.New("non-2xx status code")
	ErrEmptyMethod      = errors.New("empty http method")
	ErrEmptyUrl         = errors.New("empty url")
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
