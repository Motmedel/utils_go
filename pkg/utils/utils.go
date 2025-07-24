package utils

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

func GetConversionValue[T any](value any) (T, error) {
	convertedValue, ok := value.(T)
	if !ok {
		return convertedValue, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %T", motmedelErrors.ErrConversionNotOk, value),
			value,
		)
	}

	return convertedValue, nil
}

func GetNonZoneConversionValue[T comparable](value any) (T, error) {
	var zero T

	convertedValue, err := GetConversionValue[T](value)
	if err != nil {
		return zero, fmt.Errorf("get context value: %w", err)
	}

	if convertedValue == zero {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrContextZeroValue)
	}

	return convertedValue, nil
}

// TODO: Add `nil` check for interfaces.
