package json

import (
	"encoding/json"
	"fmt"
	"github.com/Motmedel/utils_go/pkg/errors"
	"io"
	"reflect"
	"slices"
	"strings"
)

func DecodeJson[T any](reader io.Reader) (T, error) {
	var obj T

	data, err := io.ReadAll(reader)
	if err != nil {
		return obj, &errors.CauseError{Message: "An error occurred when reading the data.", Cause: err}
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return obj, &errors.InputError{
			Message: "An error occurred when unmarshalling the data.",
			Cause:   err,
			Input:   data,
		}
	}

	return obj, err
}

func Jsonify(input any) any {
	inputValue := reflect.ValueOf(input)

	switch inputValue.Kind() {
	case reflect.Ptr:
		return Jsonify(inputValue.Elem().Interface())
	case reflect.Struct:
		structMap := make(map[string]any)
		inputValueType := inputValue.Type()
		for i := 0; i < inputValue.NumField(); i++ {
			field := inputValueType.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				continue
			}

			tag := field.Tag.Get("json")
			if tag == "-" {
				continue
			}

			value := inputValue.Field(i)

			tagSplit := strings.Split(tag, ",")
			hasOmitEmpty := slices.Contains(tagSplit, "omitempty")

			if value.IsZero() && hasOmitEmpty {
				continue
			}

			if field.Anonymous {
				if embeddedValueMap, ok := Jsonify(value.Interface()).(map[string]any); ok {
					for embeddedValueMapKey, embeddedValueMapValue := range embeddedValueMap {
						structMap[embeddedValueMapKey] = embeddedValueMapValue
					}
				} else {
					continue
				}
			} else {
				key := field.Name
				if nameIndex := slices.IndexFunc(tagSplit, func(s string) bool { return s != "omitempty" }); nameIndex >= 0 {
					key = tagSplit[nameIndex]
				}

				structMap[key] = Jsonify(value.Interface())
			}
		}
		return structMap
	case reflect.Slice, reflect.Array:
		elements := make([]any, inputValue.Len())
		for i := 0; i < inputValue.Len(); i++ {
			elements[i] = Jsonify(inputValue.Index(i).Interface())
		}
		return elements
	case reflect.Map:
		mapKeys := inputValue.MapKeys()
		m := make(map[string]any)
		for _, mapKey := range mapKeys {
			m[fmt.Sprintf("%v", mapKey)] = Jsonify(inputValue.MapIndex(mapKey).Interface())
		}
		return m
	default:
		return inputValue.Interface()
	}
}
