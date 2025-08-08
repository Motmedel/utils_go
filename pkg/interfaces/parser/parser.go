package parser

import "errors"

var ErrNilParser = errors.New("nil parser")

type Parser[T any, U any] interface {
	Parse(U) (T, error)
}

type ParserFunction[T any, U any] func(U) (T, error)

func (pf ParserFunction[T, U]) Parse(input U) (T, error) {
	return pf(input)
}
