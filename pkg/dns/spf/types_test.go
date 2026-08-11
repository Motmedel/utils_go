package spf

import (
	"encoding/json/v2"
	"reflect"
	"testing"
)

func TestRecord_OmitsTermsInJson(t *testing.T) {
	t.Parallel()

	record := &Record{
		Domain: "example.com",
		Raw:    "v=spf1 -all",
		Terms: []any{
			&Directive{
				Index:     0,
				Qualifier: "-",
				Mechanism: &Mechanism{Label: "all"},
			},
		},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	if _, ok := payload["terms"]; ok {
		t.Fatalf("terms should not be serialized: %s", string(data))
	}
	if got := payload["domain"]; got != "example.com" {
		t.Fatalf("domain: got %v want example.com", got)
	}
	if got := payload["raw"]; got != "v=spf1 -all" {
		t.Fatalf("raw: got %v want v=spf1 -all", got)
	}
}

func TestRecord_ModifiersAndDirectives(t *testing.T) {
	t.Parallel()

	directive0 := &Directive{Index: 0, Qualifier: "+", Mechanism: &Mechanism{Label: "mx"}}
	directive1 := &Directive{Index: 2, Qualifier: "-", Mechanism: &Mechanism{Label: "all"}}
	modifier := &Modifier{Index: 1, Label: "redirect", Value: "_spf.example.com"}

	record := &Record{
		Terms: []any{directive0, modifier, directive1},
	}

	if got := record.Directives(); !reflect.DeepEqual(got, []*Directive{directive0, directive1}) {
		t.Fatalf("directives mismatch: got %v", got)
	}
	if got := record.Modifiers(); !reflect.DeepEqual(got, []*Modifier{modifier}) {
		t.Fatalf("modifiers mismatch: got %v", got)
	}
}
