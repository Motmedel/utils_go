package body_parser

import (
	"net/http"

	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
)

type BodyParser[T any] interface {
	Parse(*http.Request, []byte) (T, *response_error.ResponseError)
}

type BodyParserFunction[T any] func(*http.Request, []byte) (T, *response_error.ResponseError)

func (bpf BodyParserFunction[T]) Parse(request *http.Request, body []byte) (T, *response_error.ResponseError) {
	return bpf(request, body)
}
