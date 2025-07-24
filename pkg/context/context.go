package context

import (
	"context"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type errorContextType struct{}

var ErrorContextKey errorContextType

func WithErrorContextValue(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ErrorContextKey, err)
}
func GetContextValue[T any](ctx context.Context, key any) (T, error) {
	return utils.GetConversionValue[T](ctx.Value(key))
}

func GetNonZeroContextValue[T comparable](ctx context.Context, key any) (T, error) {
	return utils.GetNonZoneConversionValue[T](ctx.Value(key))
}
