package cbor

import (
	"errors"
	"reflect"
	"testing"
)

type testEmbedded struct {
	Type      string `json:"type"`
	ProjectId string `json:"project_id,omitzero"`
}

type testDocument struct {
	FileName string `cborschema:"file_name"`
	Content  []byte `cborschema:"content"`
}

type testOrder struct {
	testEmbedded
	Name      string          `json:"name" jsonschema:"name"`
	Priority  int             `cborschema:"priority,optional"`
	Approved  bool            `json:"approved"`
	Documents []*testDocument `json:"documents,omitzero"`
	Labels    map[string]int  `json:"labels,omitzero"`
	Ignored   string          `json:"-"`
	Anything  any             `json:"anything,omitzero"`
}

func testOrderValue() map[any]any {
	return map[any]any{
		"type":       "person",
		"project_id": "p-1",
		"name":       "Meriadoc",
		"priority":   int64(3),
		"approved":   true,
		"documents": []any{
			map[any]any{"file_name": "a.pdf", "content": []byte{1, 2, 3}},
			nil,
		},
		"labels":   map[any]any{"x": int64(1)},
		"anything": "value",
	}
}

func TestUnmarshalValueStruct(t *testing.T) {
	var order testOrder
	if err := UnmarshalValue(testOrderValue(), &order); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}

	expected := testOrder{
		testEmbedded: testEmbedded{Type: "person", ProjectId: "p-1"},
		Name:         "Meriadoc",
		Priority:     3,
		Approved:     true,
		Documents: []*testDocument{
			{FileName: "a.pdf", Content: []byte{1, 2, 3}},
			nil,
		},
		Labels:   map[string]int{"x": 1},
		Anything: "value",
	}

	if !reflect.DeepEqual(order, expected) {
		t.Errorf("unmarshal value: got %#v, want %#v", order, expected)
	}
}

func TestUnmarshalRoundTrip(t *testing.T) {
	data, err := Encode(testOrderValue())
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	var order testOrder
	if err := Unmarshal(data, &order); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if order.Name != "Meriadoc" || len(order.Documents) != 2 || order.Labels["x"] != 1 {
		t.Errorf("unmarshal: unexpected result %#v", order)
	}
}

func TestUnmarshalValuePointerTarget(t *testing.T) {
	var order *testOrder
	if err := UnmarshalValue(testOrderValue(), &order); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}
	if order == nil || order.Name != "Meriadoc" {
		t.Errorf("unmarshal value: unexpected result %#v", order)
	}
}

func TestUnmarshalValueUnknownKeysIgnored(t *testing.T) {
	value := testOrderValue()
	value["unknown"] = "value"

	var order testOrder
	if err := UnmarshalValue(value, &order); err != nil {
		t.Errorf("unmarshal value: %v", err)
	}
}

func TestUnmarshalValueSkippedField(t *testing.T) {
	value := testOrderValue()
	value["-"] = "value"
	value["Ignored"] = "value"

	var order testOrder
	if err := UnmarshalValue(value, &order); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}
	if order.Ignored != "" {
		t.Errorf("expected skipped field to remain empty, got %q", order.Ignored)
	}
}

func TestUnmarshalValueErrors(t *testing.T) {
	testCases := []struct {
		name          string
		value         any
		target        any
		expectedError error
	}{
		{name: "nil target", value: "text", target: nil, expectedError: ErrInvalidTarget},
		{name: "non-pointer target", value: "text", target: "text", expectedError: ErrInvalidTarget},
		{name: "text into integer", value: "text", target: new(int), expectedError: ErrTypeMismatch},
		{name: "integer overflow", value: int64(300), target: new(int8), expectedError: ErrTypeMismatch},
		{name: "negative into unsigned", value: int64(-1), target: new(uint32), expectedError: ErrTypeMismatch},
		{name: "null into text", value: nil, target: new(string), expectedError: ErrTypeMismatch},
		{name: "array into struct", value: []any{}, target: new(testOrder), expectedError: ErrTypeMismatch},
		{name: "unsupported float target", value: int64(1), target: new(float64), expectedError: ErrUnsupportedValue},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := UnmarshalValue(testCase.value, testCase.target); !errors.Is(err, testCase.expectedError) {
				t.Errorf("expected %v, got %v", testCase.expectedError, err)
			}
		})
	}
}

func TestUnmarshalValueNullValues(t *testing.T) {
	order := testOrder{
		Documents: []*testDocument{{FileName: "a.pdf"}},
		Labels:    map[string]int{"x": 1},
	}

	value := map[any]any{"documents": nil, "labels": nil}
	if err := UnmarshalValue(value, &order); err != nil {
		t.Fatalf("unmarshal value: %v", err)
	}

	if order.Documents != nil || order.Labels != nil {
		t.Errorf("expected nils, got %#v", order)
	}
}
