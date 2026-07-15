package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// fieldTag carries the struct tag options shared with the jsonschema library's tag grammar:
// "name[,optional][,minlength:N][,maxlength:N][,minimum:N][,maximum:N][,minitems:N][,maxitems:N]
// [,format:X]", with "-" skipping the field.
type fieldTag struct {
	Name      string
	Skip      bool
	Optional  bool
	MinLength *int
	MaxLength *int
	Minimum   *int64
	Maximum   *int64
	MinItems  *int
	MaxItems  *int
	Format    string
}

func parseIntOption(key string, value string) (int, error) {
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%w: %s: %w", ErrMalformedTag, key, err)
	}
	return parsedValue, nil
}

// parseSchemaTag parses a cborschema or jsonschema tag value. Unknown options are errors when
// strict (cborschema), and ignored otherwise (jsonschema, whose producer accepts options this
// package does not know).
func parseSchemaTag(tagValue string, strict bool) (*fieldTag, error) {
	elements := strings.Split(tagValue, ",")

	if len(elements) == 1 && elements[0] == "-" {
		return &fieldTag{Skip: true}, nil
	}

	tag := fieldTag{Name: strings.TrimSpace(elements[0])}

	for _, option := range elements[1:] {
		option = strings.ToLower(strings.TrimSpace(option))
		if option == "optional" {
			tag.Optional = true
			continue
		}

		key, value, hasValue := strings.Cut(option, ":")
		if hasValue {
			switch key {
			case "format":
				tag.Format = value
				continue
			case "minlength", "maxlength", "minitems", "maxitems":
				parsedValue, err := parseIntOption(key, value)
				if err != nil {
					return nil, err
				}
				switch key {
				case "minlength":
					tag.MinLength = &parsedValue
				case "maxlength":
					tag.MaxLength = &parsedValue
				case "minitems":
					tag.MinItems = &parsedValue
				case "maxitems":
					tag.MaxItems = &parsedValue
				}
				continue
			case "minimum", "maximum":
				parsedValue, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("%w: %s: %w", ErrMalformedTag, key, err)
				}
				if key == "minimum" {
					tag.Minimum = &parsedValue
				} else {
					tag.Maximum = &parsedValue
				}
				continue
			}
		}

		if strict {
			return nil, fmt.Errorf("%w: unknown option %q", ErrMalformedTag, option)
		}
	}

	return &tag, nil
}

func parseJsonTag(tagValue string) *fieldTag {
	elements := strings.Split(tagValue, ",")

	if len(elements) == 1 && elements[0] == "-" {
		return &fieldTag{Skip: true}
	}

	tag := fieldTag{Name: elements[0]}

	for _, option := range elements[1:] {
		switch option {
		case "omitempty", "omitzero":
			tag.Optional = true
		}
	}

	return &tag
}

func fieldTagFromStructField(structField *reflect.StructField) (*fieldTag, error) {
	if tagValue, ok := structField.Tag.Lookup("cborschema"); ok {
		tag, err := parseSchemaTag(tagValue, true)
		if err != nil {
			return nil, fmt.Errorf("parse schema tag (cborschema): %w", err)
		}
		return tag, nil
	}

	if tagValue, ok := structField.Tag.Lookup("jsonschema"); ok {
		tag, err := parseSchemaTag(tagValue, false)
		if err != nil {
			return nil, fmt.Errorf("parse schema tag (jsonschema): %w", err)
		}
		return tag, nil
	}

	if tagValue, ok := structField.Tag.Lookup("json"); ok {
		return parseJsonTag(tagValue), nil
	}

	return &fieldTag{Name: structField.Name}, nil
}

// applyFieldTag applies tag constraints to a field schema. Element-level constraints (lengths,
// bounds, format) on an array field apply to its items.
func applyFieldTag(fieldSchema *Schema, tag *fieldTag) error {
	elementSchema := fieldSchema
	if fieldSchema.Type == TypeArray {
		if tag.MinItems != nil {
			fieldSchema.MinItems = tag.MinItems
		}
		if tag.MaxItems != nil {
			fieldSchema.MaxItems = tag.MaxItems
		}
		elementSchema = fieldSchema.Items
	} else if tag.MinItems != nil || tag.MaxItems != nil {
		return fmt.Errorf("%w: item bounds on non-array field", ErrMalformedTag)
	}

	if tag.MinLength != nil || tag.MaxLength != nil {
		if elementSchema == nil || (elementSchema.Type != TypeText && elementSchema.Type != TypeBytes) {
			return fmt.Errorf("%w: length bounds on non-text, non-bytes value", ErrMalformedTag)
		}
		if tag.MinLength != nil {
			elementSchema.MinLength = tag.MinLength
		}
		if tag.MaxLength != nil {
			elementSchema.MaxLength = tag.MaxLength
		}
	}

	if tag.Minimum != nil || tag.Maximum != nil {
		if elementSchema == nil || elementSchema.Type != TypeInteger {
			return fmt.Errorf("%w: value bounds on non-integer value", ErrMalformedTag)
		}
		if tag.Minimum != nil {
			elementSchema.Minimum = tag.Minimum
		}
		if tag.Maximum != nil {
			elementSchema.Maximum = tag.Maximum
		}
	}

	if tag.Format != "" {
		if elementSchema == nil || elementSchema.Type != TypeText {
			return fmt.Errorf("%w: format on non-text value", ErrMalformedTag)
		}
		elementSchema.Format = tag.Format
	}

	return nil
}

func addStructFields(structType reflect.Type, mapSchema *Schema, visiting map[reflect.Type]bool) error {
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)

		// Flatten embedded structs without an explicit name tag, mirroring encoding/json (which
		// also flattens embedded fields of unexported struct types).
		if structField.Anonymous {
			embeddedType := structField.Type
			if embeddedType.Kind() == reflect.Pointer {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct && structField.Tag.Get("json") == "" &&
				structField.Tag.Get("jsonschema") == "" && structField.Tag.Get("cborschema") == "" {
				if err := addStructFields(embeddedType, mapSchema, visiting); err != nil {
					return err
				}
				continue
			}
		}

		if structField.PkgPath != "" {
			continue
		}

		tag, err := fieldTagFromStructField(&structField)
		if err != nil {
			return fmt.Errorf("field tag (%s.%s): %w", structType, structField.Name, err)
		}
		if tag.Skip {
			continue
		}
		if tag.Name == "" {
			tag.Name = structField.Name
		}

		fieldSchema, err := schemaFromType(structField.Type, visiting)
		if err != nil {
			return fmt.Errorf("schema from type (%s.%s): %w", structType, structField.Name, err)
		}

		if err := applyFieldTag(fieldSchema, tag); err != nil {
			return fmt.Errorf("apply field tag (%s.%s): %w", structType, structField.Name, err)
		}

		mapSchema.Properties[tag.Name] = fieldSchema
		if !tag.Optional {
			mapSchema.Required = append(mapSchema.Required, tag.Name)
		}
	}

	return nil
}

func schemaFromType(t reflect.Type, visiting map[reflect.Type]bool) (*Schema, error) {
	switch t.Kind() {
	case reflect.Pointer:
		elementSchema, err := schemaFromType(t.Elem(), visiting)
		if err != nil {
			return nil, err
		}
		elementSchema.Nullable = true
		return elementSchema, nil
	case reflect.String:
		return &Schema{Type: TypeText}, nil
	case reflect.Bool:
		return &Schema{Type: TypeBoolean}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: TypeInteger}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		minimum := int64(0)
		return &Schema{Type: TypeInteger, Minimum: &minimum}, nil
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 {
			return &Schema{Type: TypeBytes}, nil
		}

		itemsSchema, err := schemaFromType(t.Elem(), visiting)
		if err != nil {
			return nil, err
		}
		return &Schema{Type: TypeArray, Items: itemsSchema}, nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("%w: map key %s", ErrUnsupportedType, t.Key())
		}

		additionalPropertiesSchema, err := schemaFromType(t.Elem(), visiting)
		if err != nil {
			return nil, err
		}
		return &Schema{Type: TypeMap, AdditionalProperties: additionalPropertiesSchema}, nil
	case reflect.Struct:
		if visiting[t] {
			return nil, fmt.Errorf("%w: recursive type %s", ErrUnsupportedType, t)
		}
		visiting[t] = true
		defer delete(visiting, t)

		mapSchema := &Schema{Type: TypeMap, Properties: map[string]*Schema{}}
		if err := addStructFields(t, mapSchema, visiting); err != nil {
			return nil, err
		}
		return mapSchema, nil
	case reflect.Interface:
		if t.NumMethod() == 0 {
			return &Schema{Type: TypeAny}, nil
		}
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, t)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, t)
	}
}

// NewFromType derives a schema from a Go type, using the cborschema struct tag, falling back to
// jsonschema (same grammar) and json (with omitempty and omitzero marking fields optional).
func NewFromType[T any]() (*Schema, error) {
	t := reflect.TypeFor[T]()
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	derivedSchema, err := schemaFromType(t, map[reflect.Type]bool{})
	if err != nil {
		return nil, fmt.Errorf("schema from type: %w", err)
	}

	return derivedSchema, nil
}
