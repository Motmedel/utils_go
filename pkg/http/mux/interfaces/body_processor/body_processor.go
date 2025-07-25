package body_processor

import "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"

type BodyProcessor[T any] interface {
	Process(any) (T, *response_error.ResponseError)
}

type BodyProcessorFunction[T any, U any] func(U) (T, *response_error.ResponseError)

func (bpf BodyProcessorFunction[T, U]) Process(input U) (T, *response_error.ResponseError) {
	return bpf(input)
}
