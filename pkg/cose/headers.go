package cose

import (
	"math"
)

// Header parameter labels from the IANA COSE Header Parameters registry.
const (
	HeaderLabelAlgorithm     int64 = 1
	HeaderLabelCritical      int64 = 2
	HeaderLabelContentType   int64 = 3
	HeaderLabelKeyIdentifier int64 = 4
	HeaderLabelIv            int64 = 5
	HeaderLabelEphemeralKey  int64 = -1
)

func toInt64(value any) (int64, bool) {
	switch typedValue := value.(type) {
	case int64:
		return typedValue, true
	case uint64:
		if typedValue <= math.MaxInt64 {
			return int64(typedValue), true
		}
	case int:
		return int64(typedValue), true
	}

	return 0, false
}

func headerValue(headerMap map[any]any, label int64) (any, bool) {
	for key, value := range headerMap {
		if intKey, ok := toInt64(key); ok && intKey == label {
			return value, true
		}
	}

	return nil, false
}

func headerBytes(headerMap map[any]any, label int64) ([]byte, bool) {
	value, ok := headerValue(headerMap, label)
	if !ok {
		return nil, false
	}

	byteValue, ok := value.([]byte)
	return byteValue, ok
}
