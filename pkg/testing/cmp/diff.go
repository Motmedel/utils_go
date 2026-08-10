package cmp

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

const maxFormatDepth = 32

// Diff compares expected and got recursively and returns a human-readable
// report of their differences, or the empty string if they are equal. Each
// difference is reported as the path to the differing value followed by the
// expected (-) and got (+) values.
//
// Values whose type has a method of the form Equal(T) bool are compared with
// that method. Comparing unexported struct fields or non-nil function values
// panics; skip them with IgnoreFields or compare a containing type with a
// Comparer option or an Equal method.
func Diff(expected any, got any, options ...Option) string {
	differ := &differ{settings: makeSettings(options), visited: make(map[visit]struct{})}
	differ.compare("", reflect.ValueOf(expected), reflect.ValueOf(got))

	return strings.Join(differ.differences, "\n")
}

// visit identifies a pair of compared references, for cycle detection.
type visit struct {
	expectedAddress uintptr
	gotAddress      uintptr
	typ             reflect.Type
}

type differ struct {
	settings    *settings
	differences []string
	visited     map[visit]struct{}
}

func (differ *differ) report(path string, expected string, got string) {
	if path == "" {
		path = "(root)"
	}

	differ.differences = append(differ.differences, fmt.Sprintf("%s:\n\t-: %s\n\t+: %s", path, expected, got))
}

func (differ *differ) reportValues(path string, expected reflect.Value, got reflect.Value) {
	differ.report(path, formatValue(expected), formatValue(got))
}

// checkVisited reports whether the pair of references has already been
// compared, recording it otherwise.
func (differ *differ) checkVisited(expected reflect.Value, got reflect.Value) bool {
	pair := visit{expectedAddress: expected.Pointer(), gotAddress: got.Pointer(), typ: expected.Type()}
	if _, ok := differ.visited[pair]; ok {
		return true
	}

	differ.visited[pair] = struct{}{}
	return false
}

func (differ *differ) compare(path string, expected reflect.Value, got reflect.Value) {
	if !expected.IsValid() || !got.IsValid() {
		if expected.IsValid() != got.IsValid() {
			differ.reportValues(path, expected, got)
		}
		return
	}

	if expected.Type() != got.Type() {
		differ.reportValues(path, expected, got)
		return
	}

	if equal, ok := differ.settings.applyComparer(expected, got); ok {
		if !equal {
			differ.reportValues(path, expected, got)
		}
		return
	}

	if equal, ok := callEqualMethod(expected, got); ok {
		if !equal {
			differ.reportValues(path, expected, got)
		}
		return
	}

	switch expected.Kind() { //nolint:exhaustive // The default case compares the remaining, primitive kinds.
	case reflect.Pointer, reflect.Interface:
		if expected.IsNil() || got.IsNil() {
			if expected.IsNil() != got.IsNil() {
				differ.reportValues(path, expected, got)
			}
			return
		}
		if expected.Kind() == reflect.Pointer && differ.checkVisited(expected, got) {
			return
		}
		differ.compare(path, expected.Elem(), got.Elem())
	case reflect.Struct:
		structType := expected.Type()
		for i := range structType.NumField() {
			field := structType.Field(i)
			if differ.settings.isIgnoredField(structType, field.Name) {
				continue
			}
			if !field.IsExported() {
				panic(fmt.Sprintf(
					"cmp: cannot compare unexported field %v.%s; skip it with IgnoreFields or compare %v with a Comparer option or an Equal method",
					structType, field.Name, structType,
				))
			}
			differ.compare(joinPath(path, field.Name), expected.Field(i), got.Field(i))
		}
	case reflect.Slice, reflect.Map:
		if differ.settings.equateEmpty && expected.Len() == 0 && got.Len() == 0 {
			return
		}
		if expected.IsNil() || got.IsNil() {
			if expected.IsNil() != got.IsNil() {
				differ.reportValues(path, expected, got)
			}
			return
		}
		if differ.checkVisited(expected, got) {
			return
		}
		if expected.Len() != got.Len() {
			differ.reportValues(path, expected, got)
			return
		}
		if expected.Kind() == reflect.Map {
			differ.compareMapEntries(path, expected, got)
		} else {
			differ.compareElements(path, expected, got)
		}
	case reflect.Array:
		differ.compareElements(path, expected, got)
	case reflect.Func:
		if expected.IsNil() && got.IsNil() {
			return
		}
		panic(fmt.Sprintf(
			"cmp: cannot compare non-nil function values of type %v; skip them with IgnoreFields or compare a containing type with a Comparer option or an Equal method",
			expected.Type(),
		))
	case reflect.Chan, reflect.UnsafePointer:
		if expected.Pointer() != got.Pointer() {
			differ.reportValues(path, expected, got)
		}
	default:
		// Bool, the integer, float and complex kinds, Uintptr and String.
		if expected.Interface() != got.Interface() {
			differ.reportValues(path, expected, got)
		}
	}
}

func (differ *differ) compareElements(path string, expected reflect.Value, got reflect.Value) {
	for i := range expected.Len() {
		differ.compare(fmt.Sprintf("%s[%d]", path, i), expected.Index(i), got.Index(i))
	}
}

func (differ *differ) compareMapEntries(path string, expected reflect.Value, got reflect.Value) {
	for _, key := range sortedMapKeys(expected) {
		entryPath := fmt.Sprintf("%s[%s]", path, formatValue(key))
		gotValue := got.MapIndex(key)
		if !gotValue.IsValid() {
			differ.report(entryPath, formatValue(expected.MapIndex(key)), "<missing>")
			continue
		}
		differ.compare(entryPath, expected.MapIndex(key), gotValue)
	}

	for _, key := range sortedMapKeys(got) {
		if !expected.MapIndex(key).IsValid() {
			differ.report(fmt.Sprintf("%s[%s]", path, formatValue(key)), "<missing>", formatValue(got.MapIndex(key)))
		}
	}
}

// callEqualMethod compares the values via a method of the form Equal(T) bool
// if the type has one. The second return value reports whether it was called.
func callEqualMethod(expected reflect.Value, got reflect.Value) (bool, bool) {
	if !expected.CanInterface() {
		return false, false
	}
	if isNilableKind(expected.Kind()) && (expected.IsNil() || got.IsNil()) {
		return false, false
	}

	method := expected.MethodByName("Equal")
	if !method.IsValid() {
		return false, false
	}

	methodType := method.Type()
	if methodType.IsVariadic() || methodType.NumIn() != 1 || methodType.NumOut() != 1 ||
		methodType.Out(0).Kind() != reflect.Bool || !got.Type().AssignableTo(methodType.In(0)) {
		return false, false
	}

	resultValues := method.Call([]reflect.Value{got})
	if len(resultValues) != 1 {
		return false, false
	}

	return resultValues[0].Bool(), true
}

func isNilableKind(kind reflect.Kind) bool {
	switch kind { //nolint:exhaustive // The remaining kinds are not nilable.
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return true
	}

	return false
}

func joinPath(path string, fieldName string) string {
	if path == "" {
		return fieldName
	}

	return path + "." + fieldName
}

func sortedMapKeys(mapValue reflect.Value) []reflect.Value {
	keys := mapValue.MapKeys()
	slices.SortFunc(keys, func(a reflect.Value, b reflect.Value) int {
		return strings.Compare(formatValue(a), formatValue(b))
	})

	return keys
}

func formatValue(value reflect.Value) string {
	return formatValueAtDepth(value, 0)
}

func formatValueAtDepth(value reflect.Value, depth int) string {
	if depth > maxFormatDepth {
		return "..."
	}

	if !value.IsValid() {
		return "<nil>"
	}

	switch value.Kind() { //nolint:exhaustive // The default case formats the remaining, primitive kinds.
	case reflect.String:
		return strconv.Quote(value.String())
	case reflect.Pointer:
		if value.IsNil() {
			return fmt.Sprintf("(%v)(nil)", value.Type())
		}
		return "&" + formatValueAtDepth(value.Elem(), depth+1)
	case reflect.Interface:
		if value.IsNil() {
			return "nil"
		}
		return formatValueAtDepth(value.Elem(), depth+1)
	case reflect.Slice:
		if value.IsNil() {
			return fmt.Sprintf("%v(nil)", value.Type())
		}
		return formatSequence(value, depth)
	case reflect.Array:
		return formatSequence(value, depth)
	case reflect.Map:
		if value.IsNil() {
			return fmt.Sprintf("%v(nil)", value.Type())
		}
		entries := make([]string, 0, value.Len())
		for _, key := range sortedMapKeys(value) {
			entries = append(
				entries,
				formatValueAtDepth(key, depth+1)+": "+formatValueAtDepth(value.MapIndex(key), depth+1),
			)
		}
		return fmt.Sprintf("%v{%s}", value.Type(), strings.Join(entries, ", "))
	case reflect.Struct:
		structType := value.Type()
		fields := make([]string, 0, structType.NumField())
		for i := range structType.NumField() {
			field := structType.Field(i)
			fieldRepresentation := "<unexported>"
			if field.IsExported() {
				fieldRepresentation = formatValueAtDepth(value.Field(i), depth+1)
			}
			fields = append(fields, field.Name+": "+fieldRepresentation)
		}
		return fmt.Sprintf("%v{%s}", structType, strings.Join(fields, ", "))
	default:
		if !value.CanInterface() {
			return "<unexported>"
		}
		return fmt.Sprintf("%#v", value.Interface())
	}
}

func formatSequence(value reflect.Value, depth int) string {
	elements := make([]string, 0, value.Len())
	for i := range value.Len() {
		elements = append(elements, formatValueAtDepth(value.Index(i), depth+1))
	}

	return fmt.Sprintf("%v{%s}", value.Type(), strings.Join(elements, ", "))
}
