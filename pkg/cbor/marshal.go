package cbor

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

func structFieldOmitted(structField *reflect.StructField) bool {
	tagValue, ok := structField.Tag.Lookup("json")
	if !ok {
		return false
	}

	for _, option := range strings.Split(tagValue, ",")[1:] {
		switch option {
		case "omitempty", "omitzero":
			return true
		}
	}

	return false
}

func valueFromStruct(structValue reflect.Value, entries map[any]any) error {
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		fieldValue := structValue.Field(i)

		if _, ok := isFlattenedEmbed(&structField); ok {
			if fieldValue.Kind() == reflect.Pointer {
				if fieldValue.IsNil() {
					continue
				}
				fieldValue = fieldValue.Elem()
			}
			if err := valueFromStruct(fieldValue, entries); err != nil {
				return err
			}
			continue
		}

		if structField.PkgPath != "" {
			continue
		}

		name, skip := fieldName(&structField)
		if skip {
			continue
		}

		if structFieldOmitted(&structField) && fieldValue.IsZero() {
			continue
		}

		entryValue, err := valueFromGo(fieldValue)
		if err != nil {
			return fmt.Errorf("field %q: %w", name, err)
		}

		entries[name] = entryValue
	}

	return nil
}

func valueFromGo(value reflect.Value) (any, error) {
	switch value.Kind() {
	case reflect.Pointer, reflect.Interface:
		if value.IsNil() {
			return nil, nil
		}
		return valueFromGo(value.Elem())
	case reflect.String:
		return value.String(), nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		unsignedValue := value.Uint()
		if unsignedValue > math.MaxInt64 {
			return nil, fmt.Errorf("%w: %d overflows int64", ErrUnsupportedValue, unsignedValue)
		}
		return int64(unsignedValue), nil
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			if value.Kind() == reflect.Array {
				data := make([]byte, value.Len())
				reflect.Copy(reflect.ValueOf(data), value)
				return data, nil
			}
			return value.Bytes(), nil
		}

		array := make([]any, value.Len())
		for i := 0; i < value.Len(); i++ {
			item, err := valueFromGo(value.Index(i))
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			array[i] = item
		}
		return array, nil
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("%w: map key %s", ErrUnsupportedValue, value.Type().Key())
		}

		entries := make(map[any]any, value.Len())
		mapIterator := value.MapRange()
		for mapIterator.Next() {
			item, err := valueFromGo(mapIterator.Value())
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", mapIterator.Key().String(), err)
			}
			entries[mapIterator.Key().String()] = item
		}
		return entries, nil
	case reflect.Struct:
		entries := make(map[any]any)
		if err := valueFromStruct(value, entries); err != nil {
			return nil, err
		}
		return entries, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedValue, value.Type())
	}
}

// MarshalValue converts a Go value into the CBOR value model. Struct fields are named via the
// cborschema, jsonschema, or json tag (in that order of precedence); fields tagged omitempty or
// omitzero are omitted when zero, and []byte becomes a byte string.
func MarshalValue(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	return valueFromGo(reflect.ValueOf(value))
}

// Marshal serializes a Go value deterministically. See MarshalValue.
func Marshal(value any) ([]byte, error) {
	modelValue, err := MarshalValue(value)
	if err != nil {
		return nil, fmt.Errorf("marshal value: %w", err)
	}

	data, err := Encode(modelValue)
	if err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return data, nil
}
