package jsonschema

import (
	jsonv2 "encoding/json/v2"
	"reflect"
	"testing"
)

type closedObject struct {
	Name string `json:"name"`
}

type openObject struct {
	//nolint:revive // The blank field is what carries what holds for the object; it is read from the type.
	_ struct{} `jsonschema:",additionalProperties:true"`

	Name string `json:"name"`
}

type explicitlyClosedObject struct {
	//nolint:revive // The blank field is what carries what holds for the object; it is read from the type.
	_ struct{} `jsonschema:",additionalProperties:false"`

	Name string `json:"name"`
}

type markerWithoutTheOption struct {
	//nolint:revive // The blank field is what carries what holds for the object; it is read from the type.
	_ struct{} `jsonschema:",optional"`

	Name string `json:"name"`
}

func TestConvertAdditionalProperties(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		targetType reflect.Type
		expected   bool
	}{
		{name: "closed without a marker", targetType: reflect.TypeFor[closedObject](), expected: false},
		{name: "open with the marker", targetType: reflect.TypeFor[openObject](), expected: true},
		{name: "closed by the marker", targetType: reflect.TypeFor[explicitlyClosedObject](), expected: false},
		{name: "closed when the marker is silent", targetType: reflect.TypeFor[markerWithoutTheOption](), expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			schemaData, err := Convert(testCase.targetType)
			if err != nil {
				t.Fatalf("convert: %v", err)
			}

			var document map[string]any
			if err := jsonv2.Unmarshal([]byte(schemaData), &document); err != nil {
				t.Fatalf("json unmarshal: %v", err)
			}

			// A struct root is emitted as a definition the document refers to by its title.
			title, _ := document["title"].(string)
			definitions, _ := document["$defs"].(map[string]any)
			definition, _ := definitions[title].(map[string]any)
			if definition == nil {
				t.Fatalf("no definition titled %q in %s", title, schemaData)
			}

			additionalProperties, ok := definition["additionalProperties"]
			if !ok {
				t.Fatalf("no additionalProperties in %s", schemaData)
			}
			if additionalProperties != testCase.expected {
				t.Errorf("additional properties: got %v, want %t", additionalProperties, testCase.expected)
			}

			// The marker is no property of the object it describes.
			properties, _ := definition["properties"].(map[string]any)
			if _, ok := properties["_"]; ok {
				t.Errorf("the marker leaked into the properties: %s", schemaData)
			}
			if len(properties) != 1 {
				t.Errorf("expected the one property, got %v", properties)
			}
		})
	}
}
