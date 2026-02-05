package comparer

import (
	"errors"
)

var ErrNilComparer = errors.New("nil comparer")

type Comparer[T comparable] interface {
	Compare(T) (bool, error)
}

type Function[T comparable] func(T) (bool, error)

func (f Function[T]) Compare(input T) (bool, error) {
	return f(input)
}

func New[T comparable](f func(T) (bool, error)) Comparer[T] {
	return Function[T](f)
}

type AnyEqualComparer[T comparable] struct {
	Values []T
}

func (c AnyEqualComparer[T]) Compare(value T) (bool, error) {
	for _, v := range c.Values {
		if v == value {
			return true, nil
		}
	}

	return false, nil
}

func NewEqualComparer[T comparable](values ...T) *AnyEqualComparer[T] {
	return &AnyEqualComparer[T]{Values: values}
}
