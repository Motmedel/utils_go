package errors

import "errors"

var (
	ErrNilHttpClient    = errors.New("nil http client")
	ErrNilHttpResponse  = errors.New("nil http response")
	ErrNon2xxStatusCode = errors.New("non-2xx status code")
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
