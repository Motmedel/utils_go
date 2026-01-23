package query_extractor

import (
	"fmt"
	"go/ast"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/query_extractor/query_extractor_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelJsonTag "github.com/Motmedel/utils_go/pkg/json/types/tag"
	motmedelReflect "github.com/Motmedel/utils_go/pkg/reflect"
	motmedelReflectErrors "github.com/Motmedel/utils_go/pkg/reflect/errors"
)

type Parser[T any] struct {
	config *query_extractor_config.Config
}

func (p *Parser[T]) Parse(request *http.Request) (T, *response_error.ResponseError) {
	var zero T

	tType := reflect.TypeOf((*T)(nil)).Elem()
	targetType := motmedelReflect.RemoveIndirection(tType)
	if targetType.Kind() != reflect.Struct {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelReflectErrors.ErrNotStruct),
		}
	}

	if request == nil {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestUrl := request.URL
	if requestUrl == nil {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestUrl),
		}
	}

	query, err := url.ParseQuery(requestUrl.RawQuery)
	if err != nil {
		return zero, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				"Malformed query.",
				nil,
			),
		}
	}

	// Allocate a new value of type T (supports struct and *struct)
	var result reflect.Value
	if tType.Kind() == reflect.Ptr {
		result = reflect.New(targetType) // *struct
	} else {
		result = reflect.New(targetType).Elem() // struct
	}
	// structVal refers to the underlying struct to populate
	var structVal reflect.Value
	if result.Kind() == reflect.Ptr {
		structVal = result.Elem()
	} else {
		structVal = result
	}

	// Keep track of known parameters and errors
	known := map[string]struct{}{}
	var parseErrs []error

	// Helpers

	setScalar := func(v reflect.Value, s string) error {
		switch v.Kind() {
		case reflect.String:
			v.SetString(s)
			return nil
		case reflect.Bool:
			if s == "" {
				v.SetBool(true)
			} else {
				b, err := strconv.ParseBool(s)
				if err != nil {
					return fmt.Errorf("invalid bool value: %q", s)
				}
				v.SetBool(b)
			}
			return nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var bitSize int
			switch v.Kind() {
			case reflect.Int8:
				bitSize = 8
			case reflect.Int16:
				bitSize = 16
			case reflect.Int32:
				bitSize = 32
			case reflect.Int64:
				bitSize = 64
			default:
				bitSize = 0 // int
			}
			n, err := strconv.ParseInt(s, 10, bitSize)
			if err != nil {
				return fmt.Errorf("invalid integer value: %q", s)
			}
			v.SetInt(n)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var bitSize int
			switch v.Kind() {
			case reflect.Uint8:
				bitSize = 8
			case reflect.Uint16:
				bitSize = 16
			case reflect.Uint32:
				bitSize = 32
			case reflect.Uint64:
				bitSize = 64
			default:
				bitSize = 0 // uint
			}
			n, err := strconv.ParseUint(s, 10, bitSize)
			if err != nil {
				return fmt.Errorf("invalid unsigned integer value: %q", s)
			}
			v.SetUint(n)
			return nil
		case reflect.Float32, reflect.Float64:
			bitSize := 64
			if v.Kind() == reflect.Float32 {
				bitSize = 32
			}
			f, err := strconv.ParseFloat(s, bitSize)
			if err != nil {
				return fmt.Errorf("invalid float value: %q", s)
			}
			v.SetFloat(f)
			return nil
		default:
			return fmt.Errorf("unsupported scalar kind: %s", v.Kind())
		}
	}

	setFromValues := func(fieldVal reflect.Value, fieldType reflect.Type, identifier string, values []string) error {
		// Pointers are not supported for fields
		if fieldType.Kind() == reflect.Ptr {
			return fmt.Errorf("pointer fields are not supported for parameter %s", identifier)
		}

		switch fieldType.Kind() {
		case reflect.Slice:
			// Special case: []byte from a single string
			if fieldType.Elem().Kind() == reflect.Uint8 {
				if len(values) != 1 {
					return fmt.Errorf("parameter %s expects a single value", identifier)
				}
				fieldVal.SetBytes([]byte(values[0]))
				return nil
			}
			slice := reflect.MakeSlice(fieldType, 0, len(values))
			for _, s := range values {
				elem := reflect.New(fieldType.Elem()).Elem()
				if err := setScalar(elem, s); err != nil {
					return fmt.Errorf("parameter %s: %w", identifier, err)
				}
				slice = reflect.Append(slice, elem)
			}
			fieldVal.Set(slice)
			return nil
		case reflect.Array:
			if len(values) != fieldType.Len() {
				return fmt.Errorf("parameter %s expects %d values", identifier, fieldType.Len())
			}
			for i := 0; i < fieldType.Len(); i++ {
				elem := fieldVal.Index(i)
				if err := setScalar(elem, values[i]); err != nil {
					return fmt.Errorf("parameter %s: %w", identifier, err)
				}
			}
			return nil
		default:
			// Scalar
			if len(values) != 1 {
				return fmt.Errorf("parameter %s expects a single value", identifier)
			}
			if err := setScalar(fieldVal, values[0]); err != nil {
				return fmt.Errorf("parameter %s: %w", identifier, err)
			}
			return nil
		}
	}

	for i := range targetType.NumField() {
		field := targetType.Field(i)

		identifier := field.Name

		if len(identifier) == 0 || !ast.IsExported(identifier) {
			continue
		}

		fieldType := field.Type
		fieldTypeKind := fieldType.Kind()

		if fieldTypeKind == reflect.Ptr {
			return zero, &response_error.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(fmt.Errorf("pointer field not supported: %s", identifier)),
			}
		}

		optional := false

		jsonTag := motmedelJsonTag.New(field.Tag.Get("json"))
		if jsonTag != nil {
			if jsonTag.Skip {
				continue
			}
			if name := jsonTag.Name; name != "" {
				identifier = name
			}
			optional = jsonTag.OmitEmpty || jsonTag.OmitZero
		}

		known[identifier] = struct{}{}

		values, ok := query[identifier]
		if !ok {
			if optional {
				continue
			}
			parseErrs = append(parseErrs, fmt.Errorf("missing parameter: %s", identifier))
			continue
		}

		if len(values) > 1 && !(fieldTypeKind == reflect.Slice || fieldTypeKind == reflect.Array) {
			parseErrs = append(parseErrs, fmt.Errorf("multiple values for parameter: %s", identifier))
			continue
		}

		targetField := structVal.Field(i)
		if err := setFromValues(targetField, fieldType, identifier, values); err != nil {
			parseErrs = append(parseErrs, err)
			continue
		}
	}

	if !p.config.AllowAdditionalParameters {
		for key := range query {
			if _, ok := known[key]; !ok {
				parseErrs = append(parseErrs, fmt.Errorf("unknown parameter: %s", key))
			}
		}
	}

	if len(parseErrs) > 0 {
		var errorStrings []string
		for _, err := range parseErrs {
			errorStrings = append(errorStrings, err.Error())
		}
		return zero, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				"Bad query.",
				map[string]interface{}{"errors": errorStrings},
			),
		}
	}

	return result.Interface().(T), nil
}

func New[T any](options ...query_extractor_config.Option) *Parser[T] {
	return &Parser[T]{config: query_extractor_config.New(options...)}
}
