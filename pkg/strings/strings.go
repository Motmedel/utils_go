package strings

import (
	"encoding"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"reflect"
	"time"
)

// NOTE: Copied from Go source code: log/slog/text_handler.go: byteSlice

func ByteSliceFromAny(a any) ([]byte, bool) {
	if bs, ok := a.([]byte); ok {
		return bs, true
	}
	// Like Printf's %s, we allow both the slice type and the byte element type to be named.
	t := reflect.TypeOf(a)
	if t != nil && t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return reflect.ValueOf(a).Bytes(), true
	}
	return nil, false
}

func MakeTextualRepresentation(value any) (string, error) {
	switch typedValue := value.(type) {
	case string:
		return typedValue, nil
	case time.Time:
		// NOTE: I would have liked ISO8601, but this is good enough.
		return typedValue.Format(time.RFC3339), nil
	default:
		if tm, ok := value.(encoding.TextMarshaler); ok {
			data, err := tm.MarshalText()
			if err != nil {
				return "", &motmedelErrors.CauseError{
					Message: "An error occurred when making a textual representation using TextMarshaler.",
					Cause:   err,
				}
			}
			return string(data), err
		}

		if bs, ok := ByteSliceFromAny(value); ok {
			return string(bs), nil
		}

		return fmt.Sprintf("%#v", value), nil
	}
}
