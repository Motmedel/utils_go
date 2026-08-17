package tag

import (
	"testing"
)

func TestNewEmpty(t *testing.T) {
	t.Parallel()

	tag, err := New("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != nil {
		t.Errorf("expected nil tag for empty input, got %+v", tag)
	}

	// Whitespace-only should also yield nil.
	tag, err = New("   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != nil {
		t.Errorf("expected nil tag for whitespace input, got %+v", tag)
	}
}

func TestNewSkip(t *testing.T) {
	t.Parallel()

	tag, err := New("-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if !tag.Skip {
		t.Error("expected Skip=true")
	}
	if tag.Name != "" {
		t.Errorf("expected empty Name on skip tag, got %q", tag.Name)
	}
}

func TestNewNameAndOptional(t *testing.T) {
	t.Parallel()

	tag, err := New("fieldName,optional")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}
	if tag.Name != "fieldName" {
		t.Errorf("Name = %q, want %q", tag.Name, "fieldName")
	}
	if !tag.Optional {
		t.Error("expected Optional=true")
	}
}

func TestNewValidationConstraints(t *testing.T) {
	t.Parallel()

	tag, err := New("f,minlength:5,maxlength:100,minimum:0.5,maximum:9.5,minitems:2,maxitems:8,format:email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}

	if tag.Name != "f" {
		t.Errorf("Name = %q, want %q", tag.Name, "f")
	}
	if tag.MinLength == nil || *tag.MinLength != 5 {
		t.Errorf("MinLength = %v, want 5", tag.MinLength)
	}
	if tag.MaxLength == nil || *tag.MaxLength != 100 {
		t.Errorf("MaxLength = %v, want 100", tag.MaxLength)
	}
	if tag.Minimum == nil || *tag.Minimum != 0.5 {
		t.Errorf("Minimum = %v, want 0.5", tag.Minimum)
	}
	if tag.Maximum == nil || *tag.Maximum != 9.5 {
		t.Errorf("Maximum = %v, want 9.5", tag.Maximum)
	}
	if tag.MinItems == nil || *tag.MinItems != 2 {
		t.Errorf("MinItems = %v, want 2", tag.MinItems)
	}
	if tag.MaxItems == nil || *tag.MaxItems != 8 {
		t.Errorf("MaxItems = %v, want 8", tag.MaxItems)
	}
	if tag.Format != "email" {
		t.Errorf("Format = %q, want %q", tag.Format, "email")
	}

	// Regression: minitems/maxitems must NOT also leak into OtherOptions.
	// Before the bug fix, those cases lacked a `continue` and their raw
	// option string was appended to OtherOptions.
	for _, opt := range tag.OtherOptions {
		if opt == "minitems:2" || opt == "maxitems:8" {
			t.Errorf("minitems/maxitems leaked into OtherOptions: %q", opt)
		}
	}
}

func TestNewInvalidNumeric(t *testing.T) {
	t.Parallel()

	cases := []string{
		"f,minlength:abc",
		"f,maxlength:abc",
		"f,minimum:abc",
		"f,maximum:abc",
		"f,minitems:abc",
		"f,maxitems:abc",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			_, err := New(input)
			if err == nil {
				t.Errorf("expected error for %q, got nil", input)
			}
		})
	}
}

func TestNewUnknownOptionFallsToOtherOptions(t *testing.T) {
	t.Parallel()

	tag, err := New("f,unknown_flag,custom:value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag == nil {
		t.Fatal("expected non-nil tag")
	}

	want := map[string]bool{"unknown_flag": true, "custom:value": true}
	got := map[string]bool{}
	for _, opt := range tag.OtherOptions {
		got[opt] = true
	}
	for k := range want {
		if !got[k] {
			t.Errorf("missing %q in OtherOptions: %v", k, tag.OtherOptions)
		}
	}
}

func TestNewAdditionalProperties(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		tagString   string
		expectError bool
		expected    *bool
	}{
		{name: "true on a blank field", tagString: ",additionalProperties:true", expected: ptr(true)},
		{name: "false on a blank field", tagString: ",additionalProperties:false", expected: ptr(false)},
		{name: "named field", tagString: "fieldName,additionalProperties:true", expected: ptr(true)},
		{name: "case insensitive", tagString: ",ADDITIONALPROPERTIES:TRUE", expected: ptr(true)},
		{name: "unset when unsaid", tagString: "fieldName,optional", expected: nil},
		{name: "not a bool", tagString: ",additionalProperties:perhaps", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tag, err := New(testCase.tagString)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected an error for %q", testCase.tagString)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tag == nil {
				t.Fatal("expected non-nil tag")
			}

			if testCase.expected == nil {
				if tag.AdditionalProperties != nil {
					t.Errorf("additional properties: got %t, want unset", *tag.AdditionalProperties)
				}
			} else {
				if tag.AdditionalProperties == nil {
					t.Fatalf("additional properties: got unset, want %t", *testCase.expected)
				}
				if *tag.AdditionalProperties != *testCase.expected {
					t.Errorf("additional properties: got %t, want %t", *tag.AdditionalProperties, *testCase.expected)
				}
			}
		})
	}
}

func ptr[T any](value T) *T {
	return &value
}
