package cmp

import (
	"fmt"
	"reflect"
)

// Option configures how Diff and CompareErr compare values.
type Option interface {
	apply(settings *settings)
}

type optionFunc func(settings *settings)

func (optionFunc optionFunc) apply(settings *settings) {
	optionFunc(settings)
}

type settings struct {
	equateEmpty   bool
	ignoredFields map[reflect.Type]map[string]struct{}
	comparers     []reflect.Value
}

func makeSettings(options []Option) *settings {
	settings := &settings{ignoredFields: make(map[reflect.Type]map[string]struct{})}

	for _, option := range options {
		option.apply(settings)
	}

	return settings
}

func (settings *settings) isIgnoredField(structType reflect.Type, fieldName string) bool {
	fieldNames, ok := settings.ignoredFields[structType]
	if !ok {
		return false
	}

	_, ok = fieldNames[fieldName]
	return ok
}

// applyComparer compares the values with the first applicable Comparer
// option. The second return value reports whether one applied.
func (settings *settings) applyComparer(expected reflect.Value, got reflect.Value) (bool, bool) {
	for _, comparer := range settings.comparers {
		argumentType := comparer.Type().In(0)
		if expected.Type().AssignableTo(argumentType) && got.Type().AssignableTo(argumentType) {
			resultValues := comparer.Call([]reflect.Value{expected, got})
			if len(resultValues) != 1 {
				return false, false
			}

			return resultValues[0].Bool(), true
		}
	}

	return false, false
}

// EquateEmpty returns an Option that treats nil slices and maps as equal to
// empty ones.
func EquateEmpty() Option {
	return optionFunc(func(settings *settings) {
		settings.equateEmpty = true
	})
}

// IgnoreFields returns an Option that skips the named fields of the struct
// type of structValue during comparison. It panics if structValue is not a
// struct or a pointer to one, or if a name does not match a field.
func IgnoreFields(structValue any, fieldNames ...string) Option {
	structType := reflect.TypeOf(structValue)
	if structType != nil && structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}
	if structType == nil || structType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("cmp: IgnoreFields: not a struct type: %T", structValue))
	}

	for _, fieldName := range fieldNames {
		if _, ok := structType.FieldByName(fieldName); !ok {
			panic(fmt.Sprintf("cmp: IgnoreFields: no field %q in %v", fieldName, structType))
		}
	}

	return optionFunc(func(settings *settings) {
		fields, ok := settings.ignoredFields[structType]
		if !ok {
			fields = make(map[string]struct{})
			settings.ignoredFields[structType] = fields
		}

		for _, fieldName := range fieldNames {
			fields[fieldName] = struct{}{}
		}
	})
}

// EquateComparable returns an Option that compares values of the given
// values' types with the == operator. Used for comparable types whose
// unexported fields would otherwise make the comparison panic. It panics if a
// value's type is not comparable.
func EquateComparable(values ...any) Option {
	var comparers []reflect.Value

	for _, value := range values {
		reflectType := reflect.TypeOf(value)
		if reflectType == nil || !reflectType.Comparable() {
			panic(fmt.Sprintf("cmp: EquateComparable: not a comparable type: %T", value))
		}

		functionType := reflect.FuncOf(
			[]reflect.Type{reflectType, reflectType},
			[]reflect.Type{reflect.TypeOf(false)},
			false,
		)
		comparers = append(
			comparers,
			reflect.MakeFunc(functionType, func(arguments []reflect.Value) []reflect.Value {
				return []reflect.Value{reflect.ValueOf(arguments[0].Equal(arguments[1]))}
			}),
		)
	}

	return optionFunc(func(settings *settings) {
		settings.comparers = append(settings.comparers, comparers...)
	})
}

// Comparer returns an Option that compares values with function, which must
// have the form func(T, T) bool. It applies wherever both compared values are
// assignable to T, taking precedence over any Equal method and the built-in
// comparison. It panics if function does not have the required form.
func Comparer(function any) Option {
	functionValue := reflect.ValueOf(function)
	if !functionValue.IsValid() || functionValue.Kind() != reflect.Func || functionValue.IsNil() {
		panic(fmt.Sprintf("cmp: Comparer: not a function: %T", function))
	}

	functionType := functionValue.Type()
	if functionType.IsVariadic() || functionType.NumIn() != 2 || functionType.NumOut() != 1 ||
		functionType.In(0) != functionType.In(1) || functionType.Out(0).Kind() != reflect.Bool {
		panic(fmt.Sprintf("cmp: Comparer: not of the form func(T, T) bool: %v", functionType))
	}

	return optionFunc(func(settings *settings) {
		settings.comparers = append(settings.comparers, functionValue)
	})
}
