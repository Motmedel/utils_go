package processor

import "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"

type Processor[T any, U any] interface {
	Process(U) (T, *response_error.ResponseError)
}

type ProcessorFunction[T any, U any] func(U) (T, *response_error.ResponseError)

func (pf ProcessorFunction[T, U]) Process(input U) (T, *response_error.ResponseError) {
	return pf(input)
}
