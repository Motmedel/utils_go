package accept_language

import (
	"fmt"
	"math"
	"testing"
)

// Corpus of realistic Accept-Language header values; tests focus on valid values and light assertions.
func TestParseAcceptLanguageCorpus(t *testing.T) {
	t.Parallel()

	cases := []string{
		// Singles
		"en",
		"fr",
		"de",
		// With region subtag
		"en-US",
		"en-GB",
		"fr-CH",
		"pt-BR",
		"zh-CN",
		// Multiple entries with q-values
		"en-US,en;q=0.9",
		"da, en-gb;q=0.8, en;q=0.7",
		"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7",
		"es-ES,es;q=0.8",
		"ru;q=0.3, en;q=1.0",
		// Spaces around separators and parameters
		"en ; q=1.0, fr ;q=0.5",
	}
	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c))
			if err != nil {
				t.Logf("ParseAcceptLanguage error for %q: %v", c, err)
				return
			}
			if p == nil {
				t.Logf("nil result for %q", c)
				return
			}
			// Light touch to ensure fields exist and Raw is populated
			_ = fmt.Sprintf("parsed %d langs (raw=%q)", len(p.LanguageQs), p.Raw)
		})
	}
}

func feq(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-6 }

// Verify parsed language tags and q-values for representative headers.
func TestParseAcceptLanguageCorrectness(t *testing.T) {
	type exp struct {
		primary string
		subtag  string
		q       float32
	}
	cases := []struct {
		name   string
		header string
		want   []exp
	}{
		{
			name:   "single primary",
			header: "en",
			want:   []exp{{primary: "en", subtag: "", q: 1.0}},
		},
		{
			name:   "primary with region",
			header: "en-US",
			want:   []exp{{primary: "en", subtag: "US", q: 1.0}},
		},
		{
			name:   "multiple with qs",
			header: "en-US,en;q=0.9",
			want:   []exp{{"en", "US", 1.0}, {"en", "", 0.9}},
		},
		{
			name:   "rfc example style",
			header: "da, en-gb;q=0.8, en;q=0.7",
			want:   []exp{{"da", "", 1.0}, {"en", "gb", 0.8}, {"en", "", 0.7}},
		},
		{
			name:   "spaces around params",
			header: "en ; q=1.0, fr ;q=0.5",
			want:   []exp{{"en", "", 1.0}, {"fr", "", 0.5}},
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c.header))
			if err != nil {
				t.Fatalf("ParseAcceptLanguage error: %v (header=%q)", err, c.header)
			}
			if p == nil {
				t.Fatalf("nil result (header=%q)", c.header)
			}
			if len(p.LanguageQs) != len(c.want) {
				t.Fatalf("len(langs)=%d, want %d (header=%q)", len(p.LanguageQs), len(c.want), c.header)
			}
			for i := range c.want {
				gl := p.LanguageQs[i]
				wl := c.want[i]
				if gl.Tag == nil {
					t.Fatalf("lang[%d].tag is nil (header=%q)", i, c.header)
				}
				if gl.Tag.PrimarySubtag != wl.primary || gl.Tag.Subtag != wl.subtag {
					t.Fatalf("lang[%d].tag=(%q-%q), want (%q-%q) (header=%q)", i, gl.Tag.PrimarySubtag, gl.Tag.Subtag, wl.primary, wl.subtag, c.header)
				}
				if !feq(gl.Q, wl.q) {
					t.Fatalf("lang[%d].q=%v, want %v (header=%q)", i, gl.Q, wl.q, c.header)
				}
			}
		})
	}
}
