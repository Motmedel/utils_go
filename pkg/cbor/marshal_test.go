package cbor

import (
	"reflect"
	"testing"
)

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	order := testOrder{
		testEmbedded: testEmbedded{Type: "person", ProjectId: "p-1"},
		Name:         "Meriadoc",
		Priority:     3,
		Approved:     true,
		Documents: []*testDocument{
			{FileName: "a.pdf", Content: []byte{1, 2, 3}},
		},
		Labels:   map[string]int{"x": 1},
		Anything: "value",
	}

	data, err := Marshal(&order)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded testOrder
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(decoded, order) {
		t.Errorf("round trip: got %#v, want %#v", decoded, order)
	}
}

func TestMarshalOmitsZeroTaggedFields(t *testing.T) {
	value, err := MarshalValue(
		&testOrder{testEmbedded: testEmbedded{Type: "person"}, Name: "Meriadoc"},
	)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}

	entries, ok := value.(map[any]any)
	if !ok {
		t.Fatalf("expected a map, got %T", value)
	}

	// project_id, documents, labels, and anything carry omitzero/omitempty and are zero.
	for _, absentKey := range []string{"project_id", "documents", "labels", "anything"} {
		if _, ok := entries[absentKey]; ok {
			t.Errorf("expected %q to be omitted, got %v", absentKey, entries[absentKey])
		}
	}

	// Fields without omission tags stay present even when zero.
	for _, presentKey := range []string{"type", "name", "priority", "approved"} {
		if _, ok := entries[presentKey]; !ok {
			t.Errorf("expected %q to be present", presentKey)
		}
	}
}

func TestMarshalUnsupportedType(t *testing.T) {
	if _, err := Marshal(struct{ Value float64 }{Value: 1.5}); err == nil {
		t.Error("expected an error for a float field")
	}
}
