package request_parser

import (
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"net/http"
)

type RequestParser[T any] interface {
	Parse(*http.Request) (T, *muxTypesResponseError.ResponseError)
}

type RequestParserFunction[T any] func(r *http.Request) (T, *muxTypesResponseError.ResponseError)

func (f RequestParserFunction[T]) Parse(request *http.Request) (T, *muxTypesResponseError.ResponseError) {
	return f(request)
}
