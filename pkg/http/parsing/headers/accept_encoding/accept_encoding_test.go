package accept_encoding

import (
	"fmt"
	"math"
	"testing"
)

// Corpus of realistic Accept-Encoding header values; tests focus on valid values and light assertions.
func TestParseAcceptEncodingCorpus(t *testing.T) {
	t.Parallel()

	cases := []string{
		"gzip",
		"br",
		"deflate",
		"zstd",
		"identity",
		"*",
		"gzip, deflate, br",
		"gzip;q=1.0, identity; q=0.5, *;q=0",
		"br;q=0.9, gzip;q=0.8, *;q=0.1",
		"zstd, br;q=0.7, gzip;q=0.3",
	}
	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c))
			if err != nil {
				t.Logf("ParseAcceptEncoding error for %q: %v", c, err)
				return
			}
			if p == nil {
				t.Logf("nil result for %q", c)
				return
			}
			fmt.Sprintf("parsed %d encodings (raw=%q)", len(p.Encodings), p.Raw)
		})
	}
}

func feq(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-6 }

// Verify parsed codings and q-values for representative headers.
func TestParseAcceptEncodingCorrectness(t *testing.T) {
	type exp struct {
		coding string
		q      float32
	}
	cases := []struct {
		name   string
		header string
		want   []exp
	}{
		{
			name:   "single coding",
			header: "gzip",
			want:   []exp{{coding: "gzip", q: 1.0}},
		},
		{
			name:   "wildcard",
			header: "*",
			want:   []exp{{coding: "*", q: 1.0}},
		},
		{
			name:   "with q",
			header: "br;q=0.9",
			want:   []exp{{coding: "br", q: 0.9}},
		},
		{
			name:   "multiple with q and spaces",
			header: "gzip;q=1.0, identity;q=0.5, *;q=0",
			want:   []exp{{"gzip", 1.0}, {"identity", 0.5}, {"*", 0.0}},
		},
		{
			name:   "browser like ordering",
			header: "br;q=0.9, gzip;q=0.8, *;q=0.1",
			want:   []exp{{"br", 0.9}, {"gzip", 0.8}, {"*", 0.1}},
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c.header))
			if err != nil {
				t.Fatalf("ParseAcceptEncoding error: %v (header=%q)", err, c.header)
			}
			if p == nil {
				t.Fatalf("nil result (header=%q)", c.header)
			}
			if len(p.Encodings) != len(c.want) {
				t.Fatalf("len(encodings)=%d, want %d (header=%q)", len(p.Encodings), len(c.want), c.header)
			}
			for i := range c.want {
				ge := p.Encodings[i]
				we := c.want[i]
				if ge.Coding != we.coding {
					t.Fatalf("enc[%d].coding=%q, want %q (header=%q)", i, ge.Coding, we.coding, c.header)
				}
				if !feq(ge.QualityValue, we.q) {
					t.Fatalf("enc[%d].q=%v, want %v (header=%q)", i, ge.QualityValue, we.q, c.header)
				}
			}
		})
	}
}
