package endpoint

import (
	"slices"
	"testing"
)

func TestExtractInlineScripts(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "inline script",
			input:    `<html><head><script>console.log("inline");</script></head></html>`,
			expected: []string{`console.log("inline");`},
		},
		{
			name:     "import map with attributes",
			input:    `<script type="importmap">{"integrity":{"/scripts/a.js":"sha384-x"}}</script>`,
			expected: []string{`{"integrity":{"/scripts/a.js":"sha384-x"}}`},
		},
		{
			name:     "script with src is ignored",
			input:    `<script defer src="/scripts/index.js" integrity="sha384-x" crossorigin="anonymous"></script>`,
			expected: nil,
		},
		{
			name:     "multiple scripts",
			input:    `<script>first</script><script src="/x.js"></script><script>second</script>`,
			expected: []string{"first", "second"},
		},
		{
			name:     "script inside a comment is ignored",
			input:    `<!-- <script>commented</script> --><script>real</script>`,
			expected: []string{"real"},
		},
		{
			name:     "quoted attribute value containing a closing angle bracket",
			input:    `<script data-x="a>b">content</script>`,
			expected: []string{"content"},
		},
		{
			name:     "case-insensitive tags",
			input:    `<SCRIPT>upper</SCRIPT>`,
			expected: []string{"upper"},
		},
		{
			name:     "unterminated script runs to the end",
			input:    `<script>unterminated`,
			expected: []string{"unterminated"},
		},
		{
			name:     "empty inline script",
			input:    `<script></script>`,
			expected: []string{""},
		},
		{
			name:     "script content containing markup",
			input:    "<script>const a = \"<!-- not a comment -->\";</script>",
			expected: []string{`const a = "<!-- not a comment -->";`},
		},
		{
			name:     "no scripts",
			input:    `<html><body><p>text</p></body></html>`,
			expected: nil,
		},
		{
			name:     "similarly named element is not a script",
			input:    `<scripting>x</scripting>`,
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			actual := extractInlineScripts([]byte(testCase.input))
			if !slices.Equal(actual, testCase.expected) {
				t.Errorf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}

func TestMakeInlineScriptHashes(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "known hash",
			input:    `<script>console.log("inline");</script>`,
			expected: []string{"sha256-L2121qypPdYD4EOJ6AR1Amd2YKHYClryjHjORJFpR7U="},
		},
		{
			name:     "duplicate scripts are deduplicated",
			input:    `<script>console.log("inline");</script><script>console.log("inline");</script>`,
			expected: []string{"sha256-L2121qypPdYD4EOJ6AR1Amd2YKHYClryjHjORJFpR7U="},
		},
		{
			name:     "no inline scripts",
			input:    `<script src="/x.js"></script>`,
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			actual := makeInlineScriptHashes([]byte(testCase.input))
			if !slices.Equal(actual, testCase.expected) {
				t.Errorf("expected %q, got %q", testCase.expected, actual)
			}
		})
	}
}

func TestNewFromDataPathInlineScriptHashes(t *testing.T) {
	t.Parallel()

	htmlEndpoint, err := NewFromDataPath(
		"/index.html",
		[]byte(`<html><head><script>console.log("inline");</script></head></html>`),
		"Sun, 02 Aug 2026 12:49:39 GMT",
		false,
		false,
	)
	if err != nil {
		t.Fatalf("new from data path: %v", err)
	}
	if htmlEndpoint == nil || htmlEndpoint.StaticContent == nil {
		t.Fatal("nil html endpoint or static content")
	}

	expected := []string{"sha256-L2121qypPdYD4EOJ6AR1Amd2YKHYClryjHjORJFpR7U="}
	if !slices.Equal(htmlEndpoint.StaticContent.InlineScriptHashes, expected) {
		t.Errorf("expected %q, got %q", expected, htmlEndpoint.StaticContent.InlineScriptHashes)
	}

	scriptEndpoint, err := NewFromDataPath(
		"/scripts/index.js",
		[]byte(`console.log("not html");`),
		"Sun, 02 Aug 2026 12:49:39 GMT",
		false,
		false,
	)
	if err != nil {
		t.Fatalf("new from data path: %v", err)
	}
	if scriptEndpoint == nil || scriptEndpoint.StaticContent == nil {
		t.Fatal("nil script endpoint or static content")
	}
	if scriptEndpoint.StaticContent.InlineScriptHashes != nil {
		t.Errorf("expected no hashes for non-html content, got %q", scriptEndpoint.StaticContent.InlineScriptHashes)
	}
}
