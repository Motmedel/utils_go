package validator

import "errors"

var ErrNilValidator = errors.New("nil validator")

type Validator[T any] interface {
	Validate(T) error
}

type ValidatorFunction[T any] func(T) error

func (vf ValidatorFunction[T]) Validate(input T) error {
	return vf(input)
}
