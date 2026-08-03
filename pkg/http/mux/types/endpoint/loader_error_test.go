package endpoint

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestNewFromDirectory_NonexistentPath(t *testing.T) {
	t.Parallel()

	if _, err := NewFromDirectory("/nonexistent-path-xyz-abc-12345/", false, false); err == nil {
		t.Fatal("expected an error for a nonexistent directory")
	}
}

func TestNewFromDataPath_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	if _, err := NewFromDataPath("/file.unsupported", []byte("x"), "", false, false); err == nil {
		t.Fatal("expected an error for an unsupported file extension")
	}
}

func zipFrom(t *testing.T, files map[string]string) *zip.Reader {
	t.Helper()

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	for name, content := range files {
		fileWriter, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("zip create %q: %v", name, err)
		}
		if _, err := fileWriter.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %q: %v", name, err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		t.Fatalf("zip new reader: %v", err)
	}
	return reader
}

func TestNewFromZip_SkipsDirectories(t *testing.T) {
	t.Parallel()

	// A trailing slash denotes a directory entry, which must be skipped.
	endpoints, err := NewFromZip(zipFrom(t, map[string]string{"subdir/": "", "a.css": "a{}"}), false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("expected 1 endpoint (directory skipped), got %d", len(endpoints))
	}
}

func TestNewFromZip_UnsupportedExtension(t *testing.T) {
	t.Parallel()

	if _, err := NewFromZip(zipFrom(t, map[string]string{"file.unsupported": "x"}), false, false); err == nil {
		t.Fatal("expected an error for an unsupported extension in the zip")
	}
}
