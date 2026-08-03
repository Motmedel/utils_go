package tag

import (
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantNil bool
		want    Tag
	}{
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
		},
		{
			name:    "whitespace only",
			input:   "   \t ",
			wantNil: true,
		},
		{
			name:  "skip marker",
			input: "-",
			want:  Tag{Skip: true},
		},
		{
			name:  "skip marker with surrounding whitespace",
			input: "  -  ",
			want:  Tag{Skip: true},
		},
		{
			name:  "name only",
			input: "field_name",
			want:  Tag{Name: "field_name"},
		},
		{
			name:  "name trimmed of surrounding whitespace",
			input: "  field_name  ",
			want:  Tag{Name: "field_name"},
		},
		{
			name:  "name with omitempty",
			input: "field,omitempty",
			want:  Tag{Name: "field", OmitEmpty: true},
		},
		{
			name:  "name with omitzero",
			input: "field,omitzero",
			want:  Tag{Name: "field", OmitZero: true},
		},
		{
			name:  "name with string",
			input: "field,string",
			want:  Tag{Name: "field", String: true},
		},
		{
			name:  "all known options",
			input: "field,omitempty,omitzero,string",
			want:  Tag{Name: "field", OmitEmpty: true, OmitZero: true, String: true},
		},
		{
			name:  "unknown options collected",
			input: "field,omitempty,custom1,custom2",
			want:  Tag{Name: "field", OmitEmpty: true, OtherOptions: []string{"custom1", "custom2"}},
		},
		{
			name:  "empty name with option (leading comma)",
			input: ",omitempty",
			want:  Tag{Name: "", OmitEmpty: true},
		},
		{
			name:  "dash with option is a named field, not skip",
			input: "-,omitempty",
			want:  Tag{Name: "-", OmitEmpty: true},
		},
		{
			name:  "duplicate option stays a single flag",
			input: "field,omitempty,omitempty",
			want:  Tag{Name: "field", OmitEmpty: true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := New(testCase.input)
			if testCase.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil tag, got nil")
			}
			if got.Name != testCase.want.Name {
				t.Fatalf("Name: expected %q, got %q", testCase.want.Name, got.Name)
			}
			if got.OmitEmpty != testCase.want.OmitEmpty {
				t.Fatalf("OmitEmpty: expected %v, got %v", testCase.want.OmitEmpty, got.OmitEmpty)
			}
			if got.OmitZero != testCase.want.OmitZero {
				t.Fatalf("OmitZero: expected %v, got %v", testCase.want.OmitZero, got.OmitZero)
			}
			if got.String != testCase.want.String {
				t.Fatalf("String: expected %v, got %v", testCase.want.String, got.String)
			}
			if got.Skip != testCase.want.Skip {
				t.Fatalf("Skip: expected %v, got %v", testCase.want.Skip, got.Skip)
			}
			if !slices.Equal(got.OtherOptions, testCase.want.OtherOptions) {
				t.Fatalf("OtherOptions: expected %v, got %v", testCase.want.OtherOptions, got.OtherOptions)
			}
		})
	}
}
