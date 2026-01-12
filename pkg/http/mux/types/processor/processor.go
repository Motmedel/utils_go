package processor

import (
	"context"

	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
)

type Processor[T any, U any] interface {
	Process(context.Context, U) (T, *response_error.ResponseError)
}

type ProcessorFunction[T any, U any] func(context.Context, U) (T, *response_error.ResponseError)

func (pf ProcessorFunction[T, U]) Process(ctx context.Context, input U) (T, *response_error.ResponseError) {
	return pf(ctx, input)
}

func New[T any, U any](f func(context.Context, U) (T, *response_error.ResponseError)) Processor[T, U] {
	return ProcessorFunction[T, U](f)
}
