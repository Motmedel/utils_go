package service_config

import (
	"testing"
)

func TestWithProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                             string
		profile                          Profile
		expectedStrictTransportSecurity  bool
		expectedApiContentSecurityPolicy bool
		expectedRobotsTxt                bool
		expectedSitemap                  bool
		expectedReporting                bool
	}{
		{
			// Nothing is said to crawlers: none of them gets past the authentication to reach what
			// a robots.txt would keep it from.
			name:                             "internal api",
			profile:                          ProfileInternalApi,
			expectedStrictTransportSecurity:  true,
			expectedApiContentSecurityPolicy: true,
		},
		{
			name:                             "public api",
			profile:                          ProfilePublicApi,
			expectedStrictTransportSecurity:  true,
			expectedApiContentSecurityPolicy: true,
			expectedRobotsTxt:                true,
		},
		{
			name:                            "internal web",
			profile:                         ProfileInternalWeb,
			expectedStrictTransportSecurity: true,
			expectedReporting:               true,
		},
		{
			name:                            "public web",
			profile:                         ProfilePublicWeb,
			expectedStrictTransportSecurity: true,
			expectedRobotsTxt:               true,
			expectedSitemap:                 true,
			expectedReporting:               true,
		},
		{
			// An unknown profile decides nothing; the service made with it reports it.
			name:    "unknown",
			profile: "something else",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			config := New(WithProfile(testCase.profile))
			if config == nil {
				t.Fatal("nil config")
			}

			if config.Profile != testCase.profile {
				t.Errorf("profile: got %q, want %q", config.Profile, testCase.profile)
			}
			if config.StrictTransportSecurity != testCase.expectedStrictTransportSecurity {
				t.Errorf(
					"strict transport security: got %t, want %t",
					config.StrictTransportSecurity,
					testCase.expectedStrictTransportSecurity,
				)
			}
			if config.ApiContentSecurityPolicy != testCase.expectedApiContentSecurityPolicy {
				t.Errorf(
					"api content security policy: got %t, want %t",
					config.ApiContentSecurityPolicy,
					testCase.expectedApiContentSecurityPolicy,
				)
			}
			if config.RobotsTxt != testCase.expectedRobotsTxt {
				t.Errorf("robots txt: got %t, want %t", config.RobotsTxt, testCase.expectedRobotsTxt)
			}
			if config.Sitemap != testCase.expectedSitemap {
				t.Errorf("sitemap: got %t, want %t", config.Sitemap, testCase.expectedSitemap)
			}
			if config.Reporting != testCase.expectedReporting {
				t.Errorf("reporting: got %t, want %t", config.Reporting, testCase.expectedReporting)
			}
		})
	}
}

// TestWithProfileIsOverriddenByLaterOptions verifies that a profile is a starting point: options are
// applied in the order they are given, so what follows one wins.
func TestWithProfileIsOverriddenByLaterOptions(t *testing.T) {
	t.Parallel()

	config := New(
		WithProfile(ProfilePublicWeb),
		WithSitemap(false),
		WithReporting(false),
	)
	if config == nil {
		t.Fatal("nil config")
	}

	if config.Sitemap {
		t.Error("the sitemap was not turned off")
	}
	if config.Reporting {
		t.Error("reporting was not turned off")
	}
	if !config.RobotsTxt || !config.StrictTransportSecurity {
		t.Error("what was not overridden was turned off along with the rest")
	}
}

// TestWithProfileBeforeProfileLoses verifies the other order: a profile applied after an option
// decides everything it decides, the option included.
func TestWithProfileBeforeProfileLoses(t *testing.T) {
	t.Parallel()

	config := New(
		WithReporting(true),
		WithProfile(ProfileInternalApi),
	)
	if config == nil {
		t.Fatal("nil config")
	}

	if config.Reporting {
		t.Error("reporting survived a profile that does not report")
	}
}

func TestProfileIsValid(t *testing.T) {
	t.Parallel()

	profiles := Profiles()
	if len(profiles) == 0 {
		t.Fatal("no profiles")
	}

	for _, profile := range profiles {
		if !profile.IsValid() {
			t.Errorf("profile %q is not valid, though it is one of the defined ones", profile)
		}
		if profile.String() != string(profile) {
			t.Errorf("string: got %q, want %q", profile.String(), string(profile))
		}
	}

	for _, profile := range []Profile{"", "something else", "public web"} {
		if profile.IsValid() {
			t.Errorf("profile %q is valid, though it is not one of the defined ones", profile)
		}
	}
}
