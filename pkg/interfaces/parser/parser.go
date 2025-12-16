package parser

import (
	"context"
	"errors"
)

var ErrNilParser = errors.New("nil parser")

type Parser[T any, U any] interface {
	Parse(U) (T, error)
}

type ParserFunction[T any, U any] func(U) (T, error)

func (pf ParserFunction[T, U]) Parse(input U) (T, error) {
	return pf(input)
}

type ParserCtx[T any, U any] interface {
	Parse(context.Context, U) (T, error)
}

type ParserCtxFunction[T any, U any] func(context.Context, U) (T, error)

func (pcf ParserCtxFunction[T, U]) Parse(ctx context.Context, input U) (T, error) {
	return pcf(ctx, input)
}

func New[T any, U any](fn func(U) (T, error)) Parser[T, U] {
	return ParserFunction[T, U](fn)
}

func NewCtx[T any, U any](fn func(context.Context, U) (T, error)) ParserCtx[T, U] {
	return ParserCtxFunction[T, U](fn)
}
