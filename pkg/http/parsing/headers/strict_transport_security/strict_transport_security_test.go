package strict_transport_security

import (
	"testing"
)

// Corpus of realistic Strict-Transport-Security header values; assertions are light.
func TestParseStrictTransportSecurityCorpus(t *testing.T) {
	t.Parallel()

	cases := []string{
		"max-age=0",
		"max-age=31536000",
		"max-age=63072000; includeSubDomains",
		"max-age=10886400;includeSubDomains",
		"includeSubDomains; max-age=31536000", // order variation
		"Max-Age=31536000; IncludeSubDomains", // mixed casing
		"max-age=\"31536000\"",                // quoted numeric value
	}

	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			p, err := Parse([]byte(c))
			if err != nil {
				// Per instructions, don't fail suite for corpus; log and continue
				t.Logf("ParseStrictTransportSecurity error for %q: %v", c, err)
				return
			}
			if p == nil {
				t.Logf("nil result for %q", c)
				return
			}
		})
	}
}

// Verify parsed fields for representative headers.
func TestParseStrictTransportSecurityCorrectness(t *testing.T) {
	t.Parallel()

	type exp struct {
		maxAge int
		incl   bool
		raw    string
	}

	cases := []struct {
		name   string
		header string
		want   exp
	}{
		{
			name:   "max-age only",
			header: "max-age=31536000",
			want:   exp{maxAge: 31536000, incl: false, raw: "max-age=31536000"},
		},
		{
			name:   "with includeSubDomains",
			header: "max-age=63072000; includeSubDomains",
			want:   exp{maxAge: 63072000, incl: true, raw: "max-age=63072000; includeSubDomains"},
		},
		{
			name:   "order variation",
			header: "includeSubDomains; max-age=10886400",
			want:   exp{maxAge: 10886400, incl: true, raw: "includeSubDomains; max-age=10886400"},
		},
		{
			name:   "quoted max-age",
			header: "max-age=\"31536000\"",
			want:   exp{maxAge: 31536000, incl: false, raw: "max-age=\"31536000\""},
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p, err := Parse([]byte(c.header))
			if err != nil {
				t.Fatalf("ParseStrictTransportSecurity error: %v (header=%q)", err, c.header)
			}
			if p == nil {
				t.Fatalf("nil result (header=%q)", c.header)
			}

			if p.MaxAga != c.want.maxAge {
				t.Fatalf("MaxAga=%d, want %d (header=%q)", p.MaxAga, c.want.maxAge, c.header)
			}
			if p.IncludeSubdomains != c.want.incl {
				t.Fatalf("IncludeSubdomains=%v, want %v (header=%q)", p.IncludeSubdomains, c.want.incl, c.header)
			}
			if p.Raw != c.want.raw {
				t.Fatalf("Raw=%q, want %q (header=%q)", p.Raw, c.want.raw, c.header)
			}
		})
	}
}
