package validator

import "errors"

var ErrNilValidator = errors.New("nil validator")

type Validator[T any] interface {
	Validate(T) error
}

type Function[T any] func(T) error

func (f Function[T]) Validate(input T) error {
	return f(input)
}

func New[T any](f func(T) error) Validator[T] {
	return Function[T](f)
}
