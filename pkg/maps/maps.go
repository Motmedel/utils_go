package maps

import motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"

func MapGetNonZero[T comparable, U comparable](m map[T]U, key T) (U, error) {
	var zero U

	if m == nil {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrNilMap)
	}

	v, ok := m[key]
	if !ok {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrNotInMap)
	}
	if v == zero {
		return zero, motmedelErrors.NewWithTrace(motmedelErrors.ErrMapZeroValue)
	}

	return v, nil
}
