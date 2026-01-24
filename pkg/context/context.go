package context

import "context"

type ErrorContextType struct{}

var ErrorContextKey ErrorContextType

func WithError(parent context.Context, err error) context.Context {
	return context.WithValue(parent, ErrorContextKey, err)
}
