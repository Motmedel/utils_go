package context

import (
	"context"
)

type errorContextType struct{}

var ErrorContextKey errorContextType

func WithErrorContextValue(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ErrorContextKey, err)
}
