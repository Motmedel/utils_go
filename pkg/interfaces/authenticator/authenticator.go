package authenticator

import "context"

type Authenticator[T any, U any] interface {
	Authenticate(context.Context, U) (T, error)
}

type Function[T any, U any] func(context.Context, U) (T, error)

func (f Function[T, U]) Authenticate(ctx context.Context, input U) (T, error) {
	return f(ctx, input)
}

func New[T any, U any](fn func(context.Context, U) (T, error)) Authenticator[T, U] {
	return Function[T, U](fn)
}
