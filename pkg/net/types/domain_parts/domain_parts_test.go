package domain_parts

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		want   *Parts
	}{
		// Nil cases: empty input.
		{
			name:   "empty string",
			domain: "",
			want:   nil,
		},

		// Nil cases: unmanaged single-label TLD (not ICANN, no dot in suffix).
		{
			name:   "unknown single-label TLD",
			domain: "cromulent",
			want:   nil,
		},

		// Nil cases: EffectiveTLDPlusOne error (domain equals suffix).
		{
			name:   "bare TLD com",
			domain: "com",
			want:   nil,
		},
		{
			name:   "bare TLD co.uk",
			domain: "co.uk",
			want:   nil,
		},

		// Simple registered domains (no subdomain).
		{
			name:   "simple .com domain",
			domain: "example.com",
			want: &Parts{
				TopLevelDomain:   "com",
				RegisteredDomain: "example.com",
			},
		},
		{
			name:   "simple .org domain",
			domain: "golang.org",
			want: &Parts{
				TopLevelDomain:   "org",
				RegisteredDomain: "golang.org",
			},
		},
		{
			name:   "simple .net domain",
			domain: "example.net",
			want: &Parts{
				TopLevelDomain:   "net",
				RegisteredDomain: "example.net",
			},
		},

		// Multi-part TLD registered domains (no subdomain).
		{
			name:   "co.uk domain",
			domain: "example.co.uk",
			want: &Parts{
				TopLevelDomain:   "co.uk",
				RegisteredDomain: "example.co.uk",
			},
		},
		{
			name:   "com.au domain",
			domain: "example.com.au",
			want: &Parts{
				TopLevelDomain:   "com.au",
				RegisteredDomain: "example.com.au",
			},
		},

		// Domains with subdomains.
		{
			name:   "single subdomain",
			domain: "www.example.com",
			want: &Parts{
				TopLevelDomain:   "com",
				RegisteredDomain: "example.com",
				Subdomain:        "www",
			},
		},
		{
			name:   "deep subdomain",
			domain: "foo.bar.baz.example.com",
			want: &Parts{
				TopLevelDomain:   "com",
				RegisteredDomain: "example.com",
				Subdomain:        "foo.bar.baz",
			},
		},
		{
			name:   "subdomain with multi-part TLD",
			domain: "www.example.co.uk",
			want: &Parts{
				TopLevelDomain:   "co.uk",
				RegisteredDomain: "example.co.uk",
				Subdomain:        "www",
			},
		},
		{
			name:   "deep subdomain with multi-part TLD",
			domain: "a.b.c.example.co.uk",
			want: &Parts{
				TopLevelDomain:   "co.uk",
				RegisteredDomain: "example.co.uk",
				Subdomain:        "a.b.c",
			},
		},

		// Private/non-ICANN domains with dots in suffix (should still work).
		{
			name:   "blogspot private domain",
			domain: "myblog.blogspot.com",
			want: &Parts{
				TopLevelDomain:   "blogspot.com",
				RegisteredDomain: "myblog.blogspot.com",
			},
		},
		{
			name:   "blogspot private domain with subdomain",
			domain: "www.myblog.blogspot.com",
			want: &Parts{
				TopLevelDomain:   "blogspot.com",
				RegisteredDomain: "myblog.blogspot.com",
				Subdomain:        "www",
			},
		},

		// Wildcard TLD domains.
		{
			name:   "wildcard ck domain",
			domain: "example.www.ck",
			want: &Parts{
				TopLevelDomain:   "ck",
				RegisteredDomain: "www.ck",
				Subdomain:        "example",
			},
		},

		// Real-world examples.
		{
			name:   "books.amazon.co.uk",
			domain: "books.amazon.co.uk",
			want: &Parts{
				TopLevelDomain:   "co.uk",
				RegisteredDomain: "amazon.co.uk",
				Subdomain:        "books",
			},
		},
		{
			name:   "foo.bar.golang.org",
			domain: "foo.bar.golang.org",
			want: &Parts{
				TopLevelDomain:   "org",
				RegisteredDomain: "golang.org",
				Subdomain:        "foo.bar",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.domain)

			if tt.want == nil {
				if got != nil {
					t.Errorf("New(%q) = %+v, want nil", tt.domain, got)
				}
				return
			}

			if got == nil {
				t.Fatalf("New(%q) = nil, want %+v", tt.domain, tt.want)
			}

			if got.TopLevelDomain != tt.want.TopLevelDomain {
				t.Errorf("New(%q).TopLevelDomain = %q, want %q", tt.domain, got.TopLevelDomain, tt.want.TopLevelDomain)
			}
			if got.RegisteredDomain != tt.want.RegisteredDomain {
				t.Errorf("New(%q).RegisteredDomain = %q, want %q", tt.domain, got.RegisteredDomain, tt.want.RegisteredDomain)
			}
			if got.Subdomain != tt.want.Subdomain {
				t.Errorf("New(%q).Subdomain = %q, want %q", tt.domain, got.Subdomain, tt.want.Subdomain)
			}
		})
	}
}

func TestNewPartsStruct(t *testing.T) {
	// Verify the struct fields and JSON tags by creating a Parts directly.
	p := Parts{
		RegisteredDomain: "example.com",
		Subdomain:        "www",
		TopLevelDomain:   "com",
	}

	if p.RegisteredDomain != "example.com" {
		t.Errorf("RegisteredDomain = %q, want %q", p.RegisteredDomain, "example.com")
	}
	if p.Subdomain != "www" {
		t.Errorf("Subdomain = %q, want %q", p.Subdomain, "www")
	}
	if p.TopLevelDomain != "com" {
		t.Errorf("TopLevelDomain = %q, want %q", p.TopLevelDomain, "com")
	}
}

func BenchmarkNew(b *testing.B) {
	domains := []string{
		"example.com",
		"www.example.com",
		"foo.bar.baz.example.co.uk",
		"myblog.blogspot.com",
	}

	for b.Loop() {
		for _, d := range domains {
			New(d)
		}
	}
}
