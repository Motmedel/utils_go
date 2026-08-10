package accept

import (
	"math"
	"reflect"
	"testing"
)

// This test feeds a corpus of valid Accept header values into the parser.
// The goal is to exercise parsing over realistic inputs; assertions are intentionally light.
func TestParseAcceptCorpus(t *testing.T) {
	t.Parallel()

	cases := []string{
		// Simple, single media types
		"text/html",
		"application/json",
		"application/xml",
		"application/xhtml+xml",
		"multipart/form-data",
		"application/x-www-form-urlencoded",
		"application/pdf;version=1.7",
		"text/markdown;charset=utf-8",
		"text/plain;format=flowed;delsp=yes",

		// Wildcards
		"*/*",
		"text/*",

		// With quality factors (q must be last in this implementation)
		"text/*;q=0.5",
		"application/xml;q=0.9, */*;q=0.1",
		"image/jpeg;q=0.2, image/png;q=0.7, image/*;q=0.5, */*;q=0.1",
		"audio/*;q=0.2, audio/basic",
		"text/html;level=1;q=0.7",

		// Multiple ranges typical of browsers
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"image/avif,image/webp,image/apng,image/*,*/*;q=0.8",

		// Structured syntax suffixes (+json, +xml) and parameters
		"application/vnd.api+json",
		"application/problem+json",
		"application/problem+xml",
		"application/hal+json",
		"application/sparql-results+json",
		"application/geo+json;charset=utf-8;q=0.9",

		// Parameters (no spaces around keys for this parser)
		"application/json;charset=utf-8",
		"text/csv;header=present",
	}

	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c))
			if err != nil {
				// We log errors but do not fail the test suite per the instructions.
				t.Logf("ParseAccept error for %q: %v", c, err)
				return
			}
			if p == nil {
				t.Logf("nil result for %q", c)
				return
			}
		})
	}
}

// Helper used by correctness tests
func floatEq(a, b float32) bool {
	return math.Abs(float64(a-b)) < 1e-6
}

// Test that parsed structures match expectations for representative headers.
func TestParseAcceptCorrectness(t *testing.T) {
	t.Parallel()

	type expRange struct {
		typeStr string
		subStr  string
		weight  float32
		params  map[string]string
	}

	cases := []struct {
		name   string
		header string
		want   []expRange
	}{
		{
			name:   "simple single",
			header: "text/html",
			want:   []expRange{{typeStr: "text", subStr: "html", weight: 1.0, params: nil}},
		},
		{
			name:   "with parameter",
			header: "application/json;charset=utf-8",
			want:   []expRange{{typeStr: "application", subStr: "json", weight: 1.0, params: map[string]string{"charset": "utf-8"}}},
		},
		{
			name:   "wildcard with q",
			header: "*/*;q=0.5",
			want:   []expRange{{typeStr: "*", subStr: "*", weight: 0.5, params: nil}},
		},
		{
			name:   "type wildcard",
			header: "text/*;q=0.2",
			want:   []expRange{{typeStr: "text", subStr: "*", weight: 0.2, params: nil}},
		},
		{
			name:   "multiple with q",
			header: "application/xml;q=0.9, */*;q=0.1",
			want: []expRange{
				{typeStr: "application", subStr: "xml", weight: 0.9, params: nil},
				{typeStr: "*", subStr: "*", weight: 0.1, params: nil},
			},
		},
		{
			name:   "multiple params",
			header: "text/plain;format=flowed;delsp=yes",
			want:   []expRange{{typeStr: "text", subStr: "plain", weight: 1.0, params: map[string]string{"format": "flowed", "delsp": "yes"}}},
		},
		{
			name:   "structured syntax suffix",
			header: "application/vnd.api+json",
			want:   []expRange{{typeStr: "application", subStr: "vnd.api+json", weight: 1.0, params: nil}},
		},
		{
			name:   "browser like",
			header: "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			want: []expRange{
				{typeStr: "text", subStr: "html", weight: 1.0, params: nil},
				{typeStr: "application", subStr: "xhtml+xml", weight: 1.0, params: nil},
				{typeStr: "application", subStr: "xml", weight: 0.9, params: nil},
				{typeStr: "*", subStr: "*", weight: 0.8, params: nil},
			},
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c.header))
			if err != nil {
				t.Fatalf("ParseAccept error: %v (header=%q)", err, c.header)
			}
			if p == nil {
				t.Fatalf("nil result (header=%q)", c.header)
			}

			if len(p.MediaRanges) != len(c.want) {
				t.Fatalf("len(mediaRanges)=%d, want %d (header=%q)", len(p.MediaRanges), len(c.want), c.header)
			}

			for i := range c.want {
				got := p.MediaRanges[i]
				exp := c.want[i]

				if got.Type != exp.typeStr || got.Subtype != exp.subStr {
					t.Fatalf("range[%d] type/subtype=%s/%s, want %s/%s (header=%q)", i, got.Type, got.Subtype, exp.typeStr, exp.subStr, c.header)
				}
				if !floatEq(got.Weight, exp.weight) {
					t.Fatalf("range[%d] weight=%v, want %v (header=%q)", i, got.Weight, exp.weight, c.header)
				}

				// Compare parameters as maps, since order is not semantically significant
				gotParams := got.GetParameterMap(false)
				if exp.params == nil && gotParams != nil {
					t.Fatalf("range[%d] params=%v, want nil (header=%q)", i, gotParams, c.header)
				}
				if exp.params != nil {
					if gotParams == nil {
						t.Fatalf("range[%d] params=nil, want %v (header=%q)", i, exp.params, c.header)
					}
					if !reflect.DeepEqual(gotParams, exp.params) {
						t.Fatalf("range[%d] params=%v, want %v (header=%q)", i, gotParams, exp.params, c.header)
					}
				}
			}
		})
	}
}
