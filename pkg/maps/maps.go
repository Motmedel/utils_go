package maps

import (
	"fmt"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

func MapGet[T comparable, U any](m map[T]U, key T) (U, error) {
	var zero U

	if m == nil {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrNilMap)
	}

	v, ok := m[key]
	if !ok {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrNotInMap)
	}

	return v, nil
}

func MapGetNonZero[T comparable, U comparable](m map[T]U, key T) (U, error) {
	var zero U

	v, err := MapGet[T, U](m, key)
	if err != nil {
		return v, fmt.Errorf("map get: %w", err)
	}
	if v == zero || utils.IsNil(v) {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrMapZeroValue)
	}

	return v, nil
}

func MapGetConvert[U any, T comparable](m map[T]any, key T) (U, error) {
	var zero U

	v, err := MapGet[T, any](m, key)
	if err != nil {
		return zero, fmt.Errorf("map get: %w", err)
	}

	cv, err := utils.Convert[U](v)
	if err != nil {
		return zero, motmedelErrors.New(fmt.Errorf("convert: %w", err), v)
	}

	return cv, nil
}

func MapGetConvertNonZero[U comparable, T comparable](m map[T]any, key T) (U, error) {
	var zero U

	v, err := MapGetConvert[U, T](m, key)
	if err != nil {
		return zero, fmt.Errorf("map get convert: %w", err)
	}
	if v == zero || utils.IsNil(v) {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrMapZeroValue)
	}

	return v, nil
}
