package types

import (
	"testing"
)

func TestMediaRangeGetFullType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		mediaType string
		subtype   string
		normalize bool
		expected  string
	}{
		{name: "plain", mediaType: "Text", subtype: "HTML", normalize: false, expected: "Text/HTML"},
		{name: "normalized", mediaType: "Text", subtype: "HTML", normalize: true, expected: "text/html"},
		{name: "empty type", mediaType: "", subtype: "json", normalize: false, expected: "*/json"},
		{name: "empty subtype", mediaType: "text", subtype: "", normalize: false, expected: "text/*"},
		{name: "both empty", mediaType: "", subtype: "", normalize: false, expected: "*/*"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mediaRange := &MediaRange{Type: testCase.mediaType, Subtype: testCase.subtype}
			if got := mediaRange.GetFullType(testCase.normalize); got != testCase.expected {
				t.Errorf("GetFullType(%v) = %q, want %q", testCase.normalize, got, testCase.expected)
			}
		})
	}
}

func TestMediaRangeGetParameterMap(t *testing.T) {
	t.Parallel()

	t.Run("nil when empty", func(t *testing.T) {
		t.Parallel()
		mediaRange := &MediaRange{}
		if got := mediaRange.GetParameterMap(false); got != nil {
			t.Errorf("GetParameterMap() = %v, want nil", got)
		}
	})

	t.Run("normalize keys and first duplicate wins", func(t *testing.T) {
		t.Parallel()
		mediaRange := &MediaRange{
			Parameters: [][2]string{{"Charset", "UTF-8"}, {"Charset", "ascii"}, {"Q", "1"}},
		}
		got := mediaRange.GetParameterMap(true)
		if got["charset"] != "UTF-8" {
			t.Errorf("charset = %q, want %q (first duplicate wins)", got["charset"], "UTF-8")
		}
		if got["q"] != "1" {
			t.Errorf("q = %q, want %q", got["q"], "1")
		}
	})

	t.Run("no normalize preserves keys", func(t *testing.T) {
		t.Parallel()
		mediaRange := &MediaRange{Parameters: [][2]string{{"Charset", "UTF-8"}}}
		got := mediaRange.GetParameterMap(false)
		if got["Charset"] != "UTF-8" {
			t.Errorf("Charset = %q, want %q", got["Charset"], "UTF-8")
		}
	})
}

func TestGetStructuredSyntaxName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		subtype   string
		normalize bool
		expected  string
	}{
		{name: "structured", subtype: "vnd.api+JSON", normalize: false, expected: "JSON"},
		{name: "structured normalized", subtype: "vnd.api+JSON", normalize: true, expected: "json"},
		{name: "no separator", subtype: "html", normalize: false, expected: ""},
		{name: "empty", subtype: "", normalize: false, expected: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mediaRange := &MediaRange{Subtype: testCase.subtype}
			if got := mediaRange.GetStructuredSyntaxName(testCase.normalize); got != testCase.expected {
				t.Errorf("GetStructuredSyntaxName(%v) = %q, want %q", testCase.normalize, got, testCase.expected)
			}
		})
	}
}

func TestServerMediaRange(t *testing.T) {
	t.Parallel()

	serverMediaRange := &ServerMediaRange{Type: "Application", Subtype: "vnd.api+json"}
	if got := serverMediaRange.GetFullType(true); got != "application/vnd.api+json" {
		t.Errorf("GetFullType(true) = %q, want %q", got, "application/vnd.api+json")
	}
	if got := serverMediaRange.GetStructuredSyntaxName(false); got != "json" {
		t.Errorf("GetStructuredSyntaxName(false) = %q, want %q", got, "json")
	}
}

func TestMediaType(t *testing.T) {
	t.Parallel()

	mediaType := &MediaType{
		Type:       "Application",
		Subtype:    "vnd.api+json",
		Parameters: [][2]string{{"Charset", "utf-8"}},
	}
	if got := mediaType.GetFullType(true); got != "application/vnd.api+json" {
		t.Errorf("GetFullType(true) = %q, want %q", got, "application/vnd.api+json")
	}
	if got := mediaType.GetStructuredSyntaxName(false); got != "json" {
		t.Errorf("GetStructuredSyntaxName(false) = %q, want %q", got, "json")
	}
	got := mediaType.GetParametersMap(true)
	if got["charset"] != "utf-8" {
		t.Errorf("charset = %q, want %q", got["charset"], "utf-8")
	}

	empty := &MediaType{}
	if got := empty.GetParametersMap(false); got != nil {
		t.Errorf("GetParametersMap() = %v, want nil", got)
	}
}

func TestContentTypeEmbeddedMediaType(t *testing.T) {
	t.Parallel()

	contentType := &ContentType{MediaType: MediaType{Type: "text", Subtype: "html"}}
	if got := contentType.GetFullType(false); got != "text/html" {
		t.Errorf("GetFullType(false) = %q, want %q", got, "text/html")
	}
}

func TestAcceptGetPriorityOrderedEncodings(t *testing.T) {
	t.Parallel()

	low := &MediaRange{Type: "text", Subtype: "plain", Weight: 0.2}
	high := &MediaRange{Type: "text", Subtype: "html", Weight: 0.9}
	mid := &MediaRange{Type: "application", Subtype: "json", Weight: 0.5}

	accept := &Accept{MediaRanges: []*MediaRange{low, high, mid}}
	ordered := accept.GetPriorityOrderedEncodings()

	if len(ordered) != 3 {
		t.Fatalf("len = %d, want 3", len(ordered))
	}
	if ordered[0] != high || ordered[1] != mid || ordered[2] != low {
		t.Errorf("ordering = [%v %v %v], want [high mid low]", ordered[0].Weight, ordered[1].Weight, ordered[2].Weight)
	}
	// The original slice must be untouched.
	if accept.MediaRanges[0] != low {
		t.Error("source slice was mutated")
	}
}

func TestAcceptEncodingGetPriorityOrderedEncodings(t *testing.T) {
	t.Parallel()

	gzip := &Encoding{Coding: "gzip", QualityValue: 1.0}
	br := &Encoding{Coding: "br", QualityValue: 0.5}
	identity := &Encoding{Coding: "identity", QualityValue: 0.1}

	acceptEncoding := &AcceptEncoding{Encodings: []*Encoding{br, identity, gzip}}
	ordered := acceptEncoding.GetPriorityOrderedEncodings()

	if len(ordered) != 3 {
		t.Fatalf("len = %d, want 3", len(ordered))
	}
	if ordered[0] != gzip || ordered[1] != br || ordered[2] != identity {
		t.Error("encodings not ordered by quality value descending")
	}
}

func TestAcceptLanguageGetPriorityOrderedLanguages(t *testing.T) {
	t.Parallel()

	en := &LanguageQ{Tag: &LanguageTag{PrimarySubtag: "en"}, Q: 1.0}
	sv := &LanguageQ{Tag: &LanguageTag{PrimarySubtag: "sv"}, Q: 0.8}
	de := &LanguageQ{Tag: &LanguageTag{PrimarySubtag: "de"}, Q: 0.3}

	acceptLanguage := &AcceptLanguage{LanguageQs: []*LanguageQ{de, en, sv}}
	ordered := acceptLanguage.GetPriorityOrderedLanguages()

	if len(ordered) != 3 {
		t.Fatalf("len = %d, want 3", len(ordered))
	}
	if ordered[0] != en || ordered[1] != sv || ordered[2] != de {
		t.Error("languages not ordered by Q descending")
	}
}

func TestAuthorizationString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		authorization *Authorization
		expected      string
	}{
		{
			name:          "empty scheme",
			authorization: &Authorization{},
			expected:      "",
		},
		{
			name:          "token68",
			authorization: &Authorization{Scheme: "Bearer", Token68: "abc123"},
			expected:      "Bearer abc123",
		},
		{
			name:          "scheme only",
			authorization: &Authorization{Scheme: "Basic"},
			expected:      "Basic",
		},
		{
			name: "token68 takes precedence over params",
			authorization: &Authorization{
				Scheme:  "Bearer",
				Token68: "abc123",
				Params:  map[string]string{"realm": "test"},
			},
			expected: "Bearer abc123",
		},
		{
			name: "params sorted, token values unquoted",
			authorization: &Authorization{
				Scheme: "Digest",
				Params: map[string]string{"algorithm": "MD5", "realm": "a b"},
			},
			expected: `Digest algorithm=MD5, realm="a b"`,
		},
		{
			name: "quote special characters",
			authorization: &Authorization{
				Scheme: "Digest",
				Params: map[string]string{"nonce": `a"b\c`},
			},
			expected: `Digest nonce="a\"b\\c"`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.authorization.String(); got != testCase.expected {
				t.Errorf("String() = %q, want %q", got, testCase.expected)
			}
		})
	}
}

func TestETagString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		etag     *ETag
		expected string
	}{
		{name: "nil", etag: nil, expected: ""},
		{name: "strong", etag: &ETag{Tag: "abc"}, expected: `"abc"`},
		{name: "weak", etag: &ETag{Weak: true, Tag: "abc"}, expected: `W/"abc"`},
		{name: "empty tag", etag: &ETag{}, expected: `""`},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.etag.String(); got != testCase.expected {
				t.Errorf("String() = %q, want %q", got, testCase.expected)
			}
		})
	}
}
