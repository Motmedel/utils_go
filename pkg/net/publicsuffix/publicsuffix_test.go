package publicsuffix

import (
	"strings"
	"testing"
)

func TestPublicSuffix(t *testing.T) {
	tests := []struct {
		domain     string
		wantSuffix string
		wantICANN  bool
	}{
		// ICANN TLDs.
		{"com", "com", true},
		{"foo.com", "com", true},
		{"bar.foo.com", "com", true},
		{"org", "org", true},
		{"foo.org", "org", true},
		{"net", "net", true},
		{"foo.net", "net", true},

		// Multi-level ICANN domains.
		{"co.uk", "co.uk", true},
		{"foo.co.uk", "co.uk", true},
		{"bar.foo.co.uk", "co.uk", true},
		{"com.au", "com.au", true},
		{"foo.com.au", "com.au", true},

		// Private domains (not ICANN).
		{"foo.blogspot.com", "blogspot.com", false},
		{"bar.foo.blogspot.com", "blogspot.com", false},

		// Wildcard rules: *.ck means any label under ck is a public suffix.
		{"www.ck", "ck", true},
		{"foo.bar.ck", "bar.ck", true},

		// Exception to wildcard: www.ck is an exception under *.ck,
		// so www.ck is not a public suffix; the suffix reverts to "ck".
		{"www.www.ck", "ck", true},

		// Unknown/unmanaged TLDs: if no rules match, the prevailing rule is "*".
		{"cromulent", "cromulent", false},
		{"foo.cromulent", "cromulent", false},

		// Wildcard *.kawasaki.jp means foo.kawasaki.jp is a public suffix.
		{"foo.kawasaki.jp", "foo.kawasaki.jp", true},
		{"bar.foo.kawasaki.jp", "foo.kawasaki.jp", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			gotSuffix, gotICANN := PublicSuffix(tt.domain)
			if gotSuffix != tt.wantSuffix {
				t.Errorf("PublicSuffix(%q) suffix = %q, want %q", tt.domain, gotSuffix, tt.wantSuffix)
			}
			if gotICANN != tt.wantICANN {
				t.Errorf("PublicSuffix(%q) icann = %v, want %v", tt.domain, gotICANN, tt.wantICANN)
			}
		})
	}
}

func TestEffectiveTLDPlusOne(t *testing.T) {
	tests := []struct {
		domain string
		want   string
		err    bool
	}{
		// Normal cases.
		{"google.com", "google.com", false},
		{"foo.google.com", "google.com", false},
		{"www.google.com", "google.com", false},
		{"amazon.co.uk", "amazon.co.uk", false},
		{"foo.amazon.co.uk", "amazon.co.uk", false},
		{"www.books.amazon.co.uk", "amazon.co.uk", false},

		// Private domains.
		{"foo.blogspot.com", "foo.blogspot.com", false},
		{"bar.foo.blogspot.com", "foo.blogspot.com", false},

		// Error cases: domain is itself a public suffix.
		{"com", "", true},
		{"co.uk", "", true},
		{"blogspot.com", "", true},

		// Error cases: empty labels.
		{".com", "", true},
		{"com.", "", true},
		{"foo..com", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got, err := EffectiveTLDPlusOne(tt.domain)
			if tt.err {
				if err == nil {
					t.Errorf("EffectiveTLDPlusOne(%q) = %q, want error", tt.domain, got)
				}
			} else {
				if err != nil {
					t.Errorf("EffectiveTLDPlusOne(%q) error: %v", tt.domain, err)
				}
				if got != tt.want {
					t.Errorf("EffectiveTLDPlusOne(%q) = %q, want %q", tt.domain, got, tt.want)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	// List should implement cookiejar.PublicSuffixList.
	ps := List.PublicSuffix("foo.co.uk")
	if ps != "co.uk" {
		t.Errorf("List.PublicSuffix(%q) = %q, want %q", "foo.co.uk", ps, "co.uk")
	}

	s := List.String()
	if !strings.Contains(s, "publicsuffix.org") {
		t.Errorf("List.String() = %q, want it to contain 'publicsuffix.org'", s)
	}
}

func TestNodeLabel(t *testing.T) {
	// Verify that the first few TLD labels can be read correctly from the data.
	for i := range uint32(numTLD) {
		label := nodeLabel(i)
		if label == "" {
			t.Errorf("nodeLabel(%d) returned empty string", i)
		}
	}
}

func TestFind(t *testing.T) {
	// Verify that well-known TLDs can be found.
	for _, tld := range []string{"com", "org", "net", "uk", "de", "jp", "au"} {
		f := find(tld, 0, numTLD)
		if f == notFound {
			t.Errorf("find(%q, 0, numTLD) returned notFound", tld)
		}
	}

	// Verify that a nonexistent TLD is not found.
	f := find("zzzzzzzzz", 0, numTLD)
	if f != notFound {
		t.Errorf("find(%q, 0, numTLD) = %d, want notFound", "zzzzzzzzz", f)
	}
}

func TestICANN(t *testing.T) {
	tests := []struct {
		domain    string
		wantICANN bool
	}{
		{"foo.com", true},
		{"foo.co.uk", true},
		{"foo.org", true},
		{"foo.blogspot.com", false},
		{"cromulent", false},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			_, gotICANN := PublicSuffix(tt.domain)
			if gotICANN != tt.wantICANN {
				t.Errorf("PublicSuffix(%q) icann = %v, want %v", tt.domain, gotICANN, tt.wantICANN)
			}
		})
	}
}

func BenchmarkPublicSuffix(b *testing.B) {
	for b.Loop() {
		PublicSuffix("www.books.amazon.co.uk")
	}
}
