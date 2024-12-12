package strings

import (
	"encoding"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"reflect"
	"strings"
	"time"
	"unicode"
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

// quote returns a shell-escaped version of the string s
func quote(s string) string {
	if s == "" {
		return "''"
	}
	// Check if the string contains unsafe characters
	if isSafe(s) {
		return s
	}
	// Use single quotes, and put single quotes into double quotes
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// isSafe checks if all characters in the string are safe
func isSafe(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
			!strings.ContainsRune("@%+=:,./-", r) {
			return false
		}
	}
	return true
}

// ShellJoin constructs a shell-quoted string from a list of tokens (ported from Python)
func ShellJoin(args []string) string {
	var quotedArgs []string
	for _, arg := range args {
		quotedArgs = append(quotedArgs, quote(arg))
	}
	return strings.Join(quotedArgs, " ")
}
