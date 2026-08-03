package header_extractor

import "testing"

func TestNew_EmptyName(t *testing.T) {
	t.Parallel()

	if _, err := New(""); err == nil {
		t.Fatal("expected an error for an empty name")
	}
}
