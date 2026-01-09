package response_checker

import (
	"net/http"
)

type ResponseChecker interface {
	Check(*http.Response, error) bool
}

type ResponseCheckerFunction func(*http.Response, error) bool

func (f ResponseCheckerFunction) Check(response *http.Response, err error) bool {
	return f(response, err)
}

func New(f func(*http.Response, error) bool) ResponseChecker {
	return ResponseCheckerFunction(f)
}
