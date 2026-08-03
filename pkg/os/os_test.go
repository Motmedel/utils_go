package os

import (
	stdOs "os"
	"path/filepath"
	"testing"
)

func TestExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	existingFile := filepath.Join(dir, "file.txt")
	if err := stdOs.WriteFile(existingFile, []byte("data"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	testCases := []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "empty path", path: "", expected: false},
		{name: "existing file", path: existingFile, expected: true},
		{name: "existing directory", path: dir, expected: true},
		{name: "missing file", path: filepath.Join(dir, "does-not-exist"), expected: false},
		{name: "missing nested path", path: filepath.Join(dir, "missing", "deeper"), expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := Exists(testCase.path); got != testCase.expected {
				t.Fatalf("Exists(%q) = %v, expected %v", testCase.path, got, testCase.expected)
			}
		})
	}
}
