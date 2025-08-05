package utils

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"reflect"
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

func GetNonZeroConversionValue[T comparable](value any) (T, error) {
	var zero T

	convertedValue, err := GetConversionValue[T](value)
	if err != nil {
		return zero, fmt.Errorf("get conversion value: %w", err)
	}

	if convertedValue == zero {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrContextZeroValue)
	}

	return convertedValue, nil
}

func IsNil(value any) bool {
	return value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
}
