package tx_caller

import (
	"context"
	"database/sql"
)

type TxCaller[T any] interface {
	Call(context.Context, *sql.Tx) (T, error)
}

type TxCallerFunction[T any] func(context.Context, *sql.Tx) (T, error)

func (f TxCallerFunction[T]) Call(ctx context.Context, tx *sql.Tx) (T, error) {
	return f(ctx, tx)
}

func New[T any](fn func(context.Context, *sql.Tx) (T, error)) TxCaller[T] {
	return TxCallerFunction[T](fn)
}
