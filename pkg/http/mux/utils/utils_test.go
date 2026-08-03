package utils

import "testing"

func TestMakeStaticContentHeaders(t *testing.T) {
	t.Parallel()

	t.Run("all empty yields no headers", func(t *testing.T) {
		t.Parallel()
		if entries := MakeStaticContentHeaders("", "", "", ""); len(entries) != 0 {
			t.Fatalf("expected no headers, got %#v", entries)
		}
	})

	t.Run("only content type", func(t *testing.T) {
		t.Parallel()
		entries := MakeStaticContentHeaders("text/html", "", "", "")
		if len(entries) != 1 || entries[0].Name != "Content-Type" || entries[0].Value != "text/html" {
			t.Fatalf("unexpected entries: %#v", entries)
		}
	})

	t.Run("all set", func(t *testing.T) {
		t.Parallel()
		entries := MakeStaticContentHeaders(
			"text/plain",
			"max-age=60",
			`"etag"`,
			"Mon, 01 Jan 2000 00:00:00 GMT",
		)

		got := map[string]string{}
		var cacheControlOverwrite bool
		for _, entry := range entries {
			got[entry.Name] = entry.Value
			if entry.Name == "Cache-Control" {
				cacheControlOverwrite = entry.Overwrite
			}
		}

		want := map[string]string{
			"Content-Type":  "text/plain",
			"Cache-Control": "max-age=60",
			"ETag":          `"etag"`,
			"Last-Modified": "Mon, 01 Jan 2000 00:00:00 GMT",
		}
		for name, value := range want {
			if got[name] != value {
				t.Errorf("header %q = %q, want %q", name, got[name], value)
			}
		}
		if !cacheControlOverwrite {
			t.Error("expected Cache-Control to be marked Overwrite")
		}
	})
}
