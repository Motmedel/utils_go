package content_type

import (
	"fmt"
	"reflect"
	"testing"
)

// Corpus of valid Content-Type header values; realistic media types and parameters.
func TestParseContentTypeCorpus(t *testing.T) {
	t.Parallel()

	cases := []string{
		"text/html",
		"application/json",
		"application/ld+json",
		"application/xml",
		"text/plain;charset=utf-8",
		"text/markdown; charset=UTF-8",
		"application/pdf;version=1.7",
		"multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
		"multipart/form-data; boundary=\"simpleBoundary\"",
		"application/problem+json",
		"application/vnd.api+json",
		"application/x-www-form-urlencoded",
	}
	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c))
			if err != nil {
				t.Logf("ParseContentType error for %q: %v", c, err)
				return
			}
			if p == nil {
				t.Logf("nil result for %q", c)
				return
			}
			fmt.Sprintf("parsed %s/%s with %d params", p.Type, p.Subtype, len(p.Parameters))
		})
	}
}

// Verify parsed type, subtype, and parameters for representative content types.
func TestParseContentTypeCorrectness(t *testing.T) {
	t.Parallel()

	type exp struct {
		typeStr string
		subStr  string
		params  map[string]string
	}
	cases := []struct {
		name   string
		header string
		want   exp
	}{
		{
			name:   "simple html",
			header: "text/html",
			want:   exp{typeStr: "text", subStr: "html", params: nil},
		},
		{
			name:   "json with charset",
			header: "application/json;charset=utf-8",
			want:   exp{typeStr: "application", subStr: "json", params: map[string]string{"charset": "utf-8"}},
		},
		{
			name:   "structured syntax suffix",
			header: "application/ld+json",
			want:   exp{typeStr: "application", subStr: "ld+json", params: nil},
		},
		{
			name:   "multipart with boundary (token)",
			header: "multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW",
			want:   exp{typeStr: "multipart", subStr: "form-data", params: map[string]string{"boundary": "----WebKitFormBoundary7MA4YWxkTrZu0gW"}},
		},
		{
			name:   "multipart with quoted boundary",
			header: "multipart/form-data; boundary=\"simpleBoundary\"",
			want:   exp{typeStr: "multipart", subStr: "form-data", params: map[string]string{"boundary": "simpleBoundary"}},
		},
		{
			name:   "pdf with version parameter",
			header: "application/pdf;version=1.7",
			want:   exp{typeStr: "application", subStr: "pdf", params: map[string]string{"version": "1.7"}},
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c.header))
			if err != nil {
				t.Fatalf("ParseContentType error: %v (header=%q)", err, c.header)
			}
			if p == nil {
				t.Fatalf("nil result (header=%q)", c.header)
			}
			if p.Type != c.want.typeStr || p.Subtype != c.want.subStr {
				t.Fatalf("type/subtype=%s/%s, want %s/%s (header=%q)", p.Type, p.Subtype, c.want.typeStr, c.want.subStr, c.header)
			}
			gotParams := p.GetParametersMap(false)
			if c.want.params == nil && gotParams != nil {
				t.Fatalf("params=%v, want nil (header=%q)", gotParams, c.header)
			}
			if c.want.params != nil {
				if gotParams == nil {
					t.Fatalf("params=nil, want %v (header=%q)", c.want.params, c.header)
				}
				if !reflect.DeepEqual(gotParams, c.want.params) {
					t.Fatalf("params=%v, want %v (header=%q)", gotParams, c.want.params, c.header)
				}
			}
		})
	}
}
