package spf

import (
	"net"
	"testing"

	parsingUtilsErrors "github.com/Motmedel/parsing_utils/pkg/errors"
	dnsTypes "github.com/Motmedel/utils_go/pkg/dns/types"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseSpfRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          []byte
		expected       *dnsTypes.SpfRecord
		expectedErrors []error
	}{
		{
			name:           "empty data",
			input:          nil,
			expected:       nil,
			expectedErrors: []error{motmedelErrors.ErrSyntaxError, parsingUtilsErrors.ErrEmptyData},
		},
		{
			name:           "syntax error",
			input:          []byte("garbage"),
			expected:       nil,
			expectedErrors: []error{motmedelErrors.ErrSyntaxError},
		},
		{
			name:  "rfc example #1",
			input: []byte("v=spf1 +mx a:colo.example.com/28 -all"),
			expected: &dnsTypes.SpfRecord{
				Terms: []any{
					&dnsTypes.SpfDirective{
						Index:     0,
						Qualifier: "+",
						Mechanism: &dnsTypes.SpfMechanism{
							Label: "mx",
						},
					},
					&dnsTypes.SpfDirective{
						Index: 1,
						Mechanism: &dnsTypes.SpfMechanism{
							Label: "a",
							Value: "colo.example.com/28",
						},
					},
					&dnsTypes.SpfDirective{
						Index:     2,
						Qualifier: "-",
						Mechanism: &dnsTypes.SpfMechanism{
							Label: "all",
						},
					},
				},
			},
		},
		{
			name:  "rfc example #2",
			input: []byte("v=spf1 +mx redirect=_spf.example.com"),
			expected: &dnsTypes.SpfRecord{
				Terms: []any{
					&dnsTypes.SpfDirective{
						Index:     0,
						Qualifier: "+",
						Mechanism: &dnsTypes.SpfMechanism{Label: "mx"},
					},
					&dnsTypes.SpfModifier{
						Index: 1,
						Label: "redirect",
						Value: "_spf.example.com",
					},
				},
			},
		},
		{
			name:  "rfc example #3",
			input: []byte("v=spf1 ?exists:_h.%{h}._l.%{l}._o.%{o}._i.%{i}._spf.%{d} ?all"),
			expected: &dnsTypes.SpfRecord{
				Terms: []any{
					&dnsTypes.SpfDirective{
						Index:     0,
						Qualifier: "?",
						Mechanism: &dnsTypes.SpfMechanism{
							Label: "exists",
							Value: `_h.%{h}._l.%{l}._o.%{o}._i.%{i}._spf.%{d}`,
						},
					},
					&dnsTypes.SpfDirective{
						Index:     1,
						Qualifier: "?",
						Mechanism: &dnsTypes.SpfMechanism{
							Label: "all",
						},
					},
				},
			},
		},
		{
			name:  "rfc example #4",
			input: []byte("v=spf1 mx -all exp=explain._spf.%{d}"),
			expected: &dnsTypes.SpfRecord{
				Terms: []any{
					&dnsTypes.SpfDirective{
						Index:     0,
						Mechanism: &dnsTypes.SpfMechanism{Label: "mx"},
					},
					&dnsTypes.SpfDirective{
						Index:     1,
						Qualifier: "-",
						Mechanism: &dnsTypes.SpfMechanism{Label: "all"},
					},
					&dnsTypes.SpfModifier{
						Index: 2,
						Label: "exp",
						Value: `explain._spf.%{d}`,
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			spfRecord, err := ParseSpfRecord(testCase.input)
			expectedErrors := testCase.expectedErrors
			if !motmedelErrors.IsAll(err, expectedErrors...) {
				t.Fatalf("expected errors: %v, got: %v", expectedErrors, err)
			}

			expected := testCase.expected
			if expected != nil {
				expected.Raw = string(testCase.input)
			}

			if diff := cmp.Diff(expected, spfRecord); diff != "" {
				t.Fatalf("struct mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}

func makeRecord(terms ...any) *dnsTypes.SpfRecord {
	return &dnsTypes.SpfRecord{Terms: terms}
}

func TestExtractIncludeValues(t *testing.T) {
	t.Parallel()

	t.Run("nil record returns nil", func(t *testing.T) {
		t.Parallel()
		if got := ExtractIncludeValues(nil); got != nil {
			t.Fatalf("got %v want nil", got)
		}
	})

	t.Run("collects include directives regardless of qualifier", func(t *testing.T) {
		t.Parallel()

		record := makeRecord(
			&dnsTypes.SpfDirective{Index: 0, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "include", Value: "a.example"}},
			&dnsTypes.SpfDirective{Index: 1, Qualifier: "-", Mechanism: &dnsTypes.SpfMechanism{Label: "include", Value: "b.example"}},
			&dnsTypes.SpfDirective{Index: 2, Qualifier: "?", Mechanism: &dnsTypes.SpfMechanism{Label: "INCLUDE", Value: "c.example"}},
			&dnsTypes.SpfDirective{Index: 3, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "mx"}},
			&dnsTypes.SpfModifier{Index: 4, Label: "include", Value: "ignored.example"},
		)

		got := ExtractIncludeValues(record)
		want := []string{"a.example", "b.example", "c.example"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestExtractRedirectValues(t *testing.T) {
	t.Parallel()

	t.Run("nil record returns nil", func(t *testing.T) {
		t.Parallel()
		if got := ExtractRedirectValues(nil); got != nil {
			t.Fatalf("got %v want nil", got)
		}
	})

	t.Run("collects redirect modifiers ignoring directives and other labels", func(t *testing.T) {
		t.Parallel()

		record := makeRecord(
			&dnsTypes.SpfModifier{Index: 0, Label: "redirect", Value: "_spf1.example"},
			&dnsTypes.SpfModifier{Index: 1, Label: "REDIRECT", Value: "_spf2.example"},
			&dnsTypes.SpfModifier{Index: 2, Label: "exp", Value: "explain"},
			&dnsTypes.SpfDirective{Index: 3, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "redirect", Value: "ignored"}},
		)

		got := ExtractRedirectValues(record)
		want := []string{"_spf1.example", "_spf2.example"}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestExtractNetworks(t *testing.T) {
	t.Parallel()

	t.Run("nil record returns nil", func(t *testing.T) {
		t.Parallel()
		if got := ExtractNetworks(nil, false); got != nil {
			t.Fatalf("got %v want nil", got)
		}
	})

	t.Run("returns ip4 and ip6 networks", func(t *testing.T) {
		t.Parallel()

		record := makeRecord(
			&dnsTypes.SpfDirective{Index: 0, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "192.0.2.0/24"}},
			&dnsTypes.SpfDirective{Index: 1, Qualifier: "-", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "198.51.100.5"}},
			&dnsTypes.SpfDirective{Index: 2, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "IP6", Value: "2001:db8::/32"}},
			&dnsTypes.SpfDirective{Index: 3, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "mx"}},
		)

		got := ExtractNetworks(record, false)
		var gotStrings []string
		for _, n := range got {
			gotStrings = append(gotStrings, n.String())
		}
		want := []string{"192.0.2.0/24", "198.51.100.5/32", "2001:db8::/32"}
		if diff := cmp.Diff(want, gotStrings, cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("network strings mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("passOnly skips non-pass qualifiers", func(t *testing.T) {
		t.Parallel()

		record := makeRecord(
			&dnsTypes.SpfDirective{Index: 0, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "192.0.2.0/24"}},
			&dnsTypes.SpfDirective{Index: 1, Qualifier: "-", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "198.51.100.0/24"}},
			&dnsTypes.SpfDirective{Index: 2, Qualifier: "", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "203.0.113.0/24"}},
		)

		got := ExtractNetworks(record, true)
		var gotStrings []string
		for _, n := range got {
			gotStrings = append(gotStrings, n.String())
		}
		want := []string{"192.0.2.0/24", "203.0.113.0/24"}
		if diff := cmp.Diff(want, gotStrings); diff != "" {
			t.Fatalf("passOnly network strings mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid network strings are dropped", func(t *testing.T) {
		t.Parallel()

		record := makeRecord(
			&dnsTypes.SpfDirective{Index: 0, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "garbage"}},
			&dnsTypes.SpfDirective{Index: 1, Qualifier: "+", Mechanism: &dnsTypes.SpfMechanism{Label: "ip4", Value: "192.0.2.0/24"}},
		)

		got := ExtractNetworks(record, false)
		if len(got) != 1 {
			t.Fatalf("got %d networks want 1", len(got))
		}
		if _, ok := any(got[0]).(*net.IPNet); !ok {
			t.Fatalf("expected *net.IPNet, got %T", got[0])
		}
		if got[0].String() != "192.0.2.0/24" {
			t.Fatalf("got %q want %q", got[0].String(), "192.0.2.0/24")
		}
	})
}
