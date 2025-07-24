package context

import (
	"context"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

type errorContextType struct{}

var ErrorContextKey errorContextType

func WithErrorContextValue(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, ErrorContextKey, err)
}
func GetContextValue[T any](ctx context.Context, key any) (T, error) {
	var value T

	extractedValue := ctx.Value(key)
	value, ok := extractedValue.(T)
	if !ok {
		return value, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %T", motmedelErrors.ErrConversionNotOk, extractedValue),
			extractedValue,
		)
	}

	return value, nil
}

func GetNonZeroContextValue[T comparable](ctx context.Context, key any) (T, error) {
	var zeroValue T

	value, err := GetContextValue[T](ctx, key)
	if err != nil {
		return zeroValue, fmt.Errorf("get context value: %w", err)
	}

	if value == zeroValue {
		return zeroValue, motmedelErrors.NewWithTrace(motmedelErrors.ErrContextZeroValue)
	}

	return value, nil
}
