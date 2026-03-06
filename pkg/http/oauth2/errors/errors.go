package errors

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyClientId     = errors.New("empty client id")
	ErrEmptyTokenUrl     = errors.New("empty token url")
	ErrEmptyAuthUrl      = errors.New("empty auth url")
	ErrEmptyAccessToken  = errors.New("empty access token")
	ErrEmptyRefreshToken = errors.New("empty refresh token")
	ErrTokenExpired      = errors.New("token expired")
	ErrRetrieveToken     = errors.New("retrieve token")
)

type RetrieveError struct {
	StatusCode       int
	Body             []byte
	ErrorCode        string
	ErrorDescription string
	ErrorURI         string
}

func (e *RetrieveError) Is(target error) bool {
	return target == ErrRetrieveToken
}

func (e *RetrieveError) Error() string {
	if e.ErrorCode != "" {
		s := fmt.Sprintf("oauth2: %q", e.ErrorCode)
		if e.ErrorDescription != "" {
			s += fmt.Sprintf(" %q", e.ErrorDescription)
		}
		if e.ErrorURI != "" {
			s += fmt.Sprintf(" %q", e.ErrorURI)
		}
		return s
	}
	return fmt.Sprintf("oauth2: cannot fetch token: %d", e.StatusCode)
}
