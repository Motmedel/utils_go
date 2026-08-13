package matchfinder

import (
	"strings"
	"testing"
)

// The vendored implementation is covered thoroughly upstream; this exercises
// basic match finding on repetitive input.
func TestVendoredFindMatches(t *testing.T) {
	t.Parallel()

	var finder M4
	matches := finder.FindMatches(nil, []byte(strings.Repeat("abcdefgh", 64)))
	if len(matches) == 0 {
		t.Error("expected matches in repetitive input")
	}
}
