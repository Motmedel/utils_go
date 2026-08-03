package url_allower_config

import (
	"slices"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	defaults := New()
	if defaults.AllowLocalhost {
		t.Error("expected AllowLocalhost to default to false")
	}
	if len(defaults.AllowedDomains) != 0 || len(defaults.AllowedRegisteredDomains) != 0 {
		t.Errorf("expected empty domain lists by default, got %#v", defaults)
	}

	config := New(
		WithAllowLocalhost(true),
		WithAllowedDomains([]string{"a.example.com"}),
		WithAllowedRegisteredDomains([]string{"example.com"}),
	)
	if !config.AllowLocalhost {
		t.Error("expected AllowLocalhost to be true")
	}
	if !slices.Equal(config.AllowedDomains, []string{"a.example.com"}) {
		t.Errorf("allowed domains = %#v", config.AllowedDomains)
	}
	if !slices.Equal(config.AllowedRegisteredDomains, []string{"example.com"}) {
		t.Errorf("allowed registered domains = %#v", config.AllowedRegisteredDomains)
	}
}
