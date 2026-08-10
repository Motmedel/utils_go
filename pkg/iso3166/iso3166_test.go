package iso3166

import (
	"strings"
	"testing"
)

func TestCountryName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		alpha2   string
		expected string
	}{
		{name: "sweden", alpha2: "SE", expected: "Sweden"},
		{name: "united states", alpha2: "US", expected: "United States"},
		{name: "lowercase", alpha2: "se", expected: "Sweden"},
		{name: "deprecated code", alpha2: "UK", expected: "United Kingdom"},
		{name: "unknown region", alpha2: "ZZ", expected: "Unknown Region"},
		{name: "invalid", alpha2: "0X", expected: ""},
		{name: "empty", alpha2: "", expected: ""},
		{name: "three letters", alpha2: "SWE", expected: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if name := CountryName(testCase.alpha2); name != testCase.expected {
				t.Errorf("CountryName(%q) = %q, expected %q", testCase.alpha2, name, testCase.expected)
			}
		})
	}
}

func TestCountryNamesNonEmpty(t *testing.T) {
	t.Parallel()

	if len(countryNames) < 250 {
		t.Errorf("unexpectedly few entries: %d", len(countryNames))
	}

	for code, name := range countryNames {
		if len(code) != 2 || code != strings.ToUpper(code) {
			t.Errorf("malformed code %q", code)
		}
		if name == "" {
			t.Errorf("empty name for %q", code)
		}
	}
}
