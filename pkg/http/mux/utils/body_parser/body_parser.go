package body_parser

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"net/http"
)

type BodyParser[T any] interface {
	Parse(*http.Request, []byte) (any, *response_error.ResponseError)
}

type BodyParserFunction[T any] func(*http.Request, []byte) (*T, *response_error.ResponseError)

func (bpf BodyParserFunction[T]) Parse(request *http.Request, body []byte) (any, *response_error.ResponseError) {
	return bpf(request, body)
}

type BodyProcessor[T any] interface {
	Process(any) (T, *response_error.ResponseError)
}

type BodyProcessorFunction[T any] func(any) (T, *response_error.ResponseError)

func (bpf BodyProcessorFunction[T]) Process(input any) (T, *response_error.ResponseError) {
	return bpf(input)
}
