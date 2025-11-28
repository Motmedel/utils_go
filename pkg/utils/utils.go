package utils

import (
	"fmt"
	"reflect"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

func Convert[T any](value any) (T, error) {
	convertedValue, ok := value.(T)
	if !ok {
		return convertedValue, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %T", motmedelErrors.ErrConversionNotOk, value),
			value,
		)
	}

	return convertedValue, nil
}

func ConvertSlice[T any](value any) ([]T, error) {
	var zero []T

	if typedValue, ok := value.([]T); ok {
		return typedValue, nil
	}

	if anySlice, ok := value.([]any); ok {
		typedSlice := make([]T, len(anySlice))
		for i, sliceValue := range anySlice {
			typedSliceValue, ok := sliceValue.(T)
			if !ok {
				return nil, motmedelErrors.NewWithTrace(
					fmt.Errorf("%w: %T (slice value)", motmedelErrors.ErrConversionNotOk, value),
					value,
				)
			}
			typedSlice[i] = typedSliceValue
		}
		return typedSlice, nil
	}

	return zero, motmedelErrors.NewWithTrace(
		fmt.Errorf("%w: %T", motmedelErrors.ErrConversionNotOk, value),
		value,
	)
}

func ConvertToNonZero[T comparable](value any) (T, error) {
	var zero T

	convertedValue, err := Convert[T](value)
	if err != nil {
		return zero, fmt.Errorf("convert: %w", err)
	}

	if convertedValue == zero {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrZeroValue)
	}

	return convertedValue, nil
}

func IsNil(value any) bool {
	return value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil())
}
