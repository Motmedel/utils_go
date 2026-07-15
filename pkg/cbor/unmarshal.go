package cbor

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrInvalidTarget = errors.New("invalid target")
	ErrTypeMismatch  = errors.New("type mismatch")
)

// fieldName resolves the map key for a struct field from the cborschema, jsonschema, or json
// struct tag (whichever appears first), reporting whether the field is skipped ("-").
func fieldName(structField *reflect.StructField) (string, bool) {
	for _, tagKey := range []string{"cborschema", "jsonschema", "json"} {
		tagValue, ok := structField.Tag.Lookup(tagKey)
		if !ok {
			continue
		}

		if tagValue == "-" {
			return "", true
		}

		name, _, _ := strings.Cut(tagValue, ",")
		if name == "" {
			name = structField.Name
		}

		return name, false
	}

	return structField.Name, false
}

func isFlattenedEmbed(structField *reflect.StructField) (reflect.Type, bool) {
	if !structField.Anonymous {
		return nil, false
	}

	embeddedType := structField.Type
	if embeddedType.Kind() == reflect.Pointer {
		embeddedType = embeddedType.Elem()
	}

	if embeddedType.Kind() != reflect.Struct {
		return nil, false
	}

	if structField.Tag.Get("json") != "" || structField.Tag.Get("jsonschema") != "" ||
		structField.Tag.Get("cborschema") != "" {
		return nil, false
	}

	return embeddedType, true
}

func assignStructFields(entries map[any]any, target reflect.Value) error {
	structType := target.Type()

	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		fieldValue := target.Field(i)

		if _, ok := isFlattenedEmbed(&structField); ok {
			if fieldValue.Kind() == reflect.Pointer {
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}
				fieldValue = fieldValue.Elem()
			}
			if err := assignStructFields(entries, fieldValue); err != nil {
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

		entryValue, ok := entries[name]
		if !ok {
			continue
		}

		if err := assignValue(entryValue, fieldValue); err != nil {
			return fmt.Errorf("field %q: %w", name, err)
		}
	}

	return nil
}

func assignValue(value any, target reflect.Value) error {
	switch target.Kind() {
	case reflect.Pointer:
		if value == nil {
			target.SetZero()
			return nil
		}
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return assignValue(value, target.Elem())
	case reflect.Interface:
		if target.NumMethod() != 0 {
			return fmt.Errorf("%w: %s", ErrUnsupportedValue, target.Type())
		}
		if value == nil {
			target.SetZero()
		} else {
			target.Set(reflect.ValueOf(value))
		}
		return nil
	case reflect.String:
		textValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("%w: expected text, got %T", ErrTypeMismatch, value)
		}
		target.SetString(textValue)
		return nil
	case reflect.Bool:
		boolValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%w: expected boolean, got %T", ErrTypeMismatch, value)
		}
		target.SetBool(boolValue)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, ok := value.(int64)
		if !ok {
			return fmt.Errorf("%w: expected integer, got %T", ErrTypeMismatch, value)
		}
		if target.OverflowInt(intValue) {
			return fmt.Errorf("%w: %d overflows %s", ErrTypeMismatch, intValue, target.Type())
		}
		target.SetInt(intValue)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		intValue, ok := value.(int64)
		if !ok {
			return fmt.Errorf("%w: expected integer, got %T", ErrTypeMismatch, value)
		}
		if intValue < 0 || target.OverflowUint(uint64(intValue)) {
			return fmt.Errorf("%w: %d overflows %s", ErrTypeMismatch, intValue, target.Type())
		}
		target.SetUint(uint64(intValue))
		return nil
	case reflect.Slice:
		if value == nil {
			target.SetZero()
			return nil
		}

		if target.Type().Elem().Kind() == reflect.Uint8 {
			byteValue, ok := value.([]byte)
			if !ok {
				return fmt.Errorf("%w: expected bytes, got %T", ErrTypeMismatch, value)
			}
			target.SetBytes(byteValue)
			return nil
		}

		arrayValue, ok := value.([]any)
		if !ok {
			return fmt.Errorf("%w: expected array, got %T", ErrTypeMismatch, value)
		}

		slice := reflect.MakeSlice(target.Type(), len(arrayValue), len(arrayValue))
		for i, item := range arrayValue {
			if err := assignValue(item, slice.Index(i)); err != nil {
				return fmt.Errorf("index %d: %w", i, err)
			}
		}
		target.Set(slice)
		return nil
	case reflect.Map:
		if target.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("%w: map key %s", ErrUnsupportedValue, target.Type().Key())
		}

		if value == nil {
			target.SetZero()
			return nil
		}

		mapValue, ok := value.(map[any]any)
		if !ok {
			return fmt.Errorf("%w: expected map, got %T", ErrTypeMismatch, value)
		}

		targetMap := reflect.MakeMapWithSize(target.Type(), len(mapValue))
		elementType := target.Type().Elem()
		keyType := target.Type().Key()

		for key, item := range mapValue {
			textKey, ok := key.(string)
			if !ok {
				return fmt.Errorf("%w: expected text key, got %T", ErrTypeMismatch, key)
			}

			element := reflect.New(elementType).Elem()
			if err := assignValue(item, element); err != nil {
				return fmt.Errorf("key %q: %w", textKey, err)
			}

			targetMap.SetMapIndex(reflect.ValueOf(textKey).Convert(keyType), element)
		}
		target.Set(targetMap)
		return nil
	case reflect.Struct:
		mapValue, ok := value.(map[any]any)
		if !ok {
			return fmt.Errorf("%w: expected map, got %T", ErrTypeMismatch, value)
		}
		return assignStructFields(mapValue, target)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedValue, target.Type())
	}
}

// UnmarshalValue maps a decoded CBOR value into target, which must be a non-nil pointer. Struct
// fields are matched by the cborschema, jsonschema, or json tag name (in that order of
// precedence), falling back to the field name; unknown map keys are ignored.
func UnmarshalValue(value any, target any) error {
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return ErrInvalidTarget
	}

	return assignValue(value, targetValue.Elem())
}

// Unmarshal decodes data and maps the resulting value into target. See UnmarshalValue.
func Unmarshal(data []byte, target any) error {
	value, err := Decode(data)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	return UnmarshalValue(value, target)
}
