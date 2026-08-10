// Package iso3166 maps ISO 3166-1 alpha-2 country codes to their English names. Deprecated
// codes resolve to the name of their canonical replacement (e.g. "UK" to "United Kingdom").
package iso3166

import "strings"

// CountryName returns the English name of the country with the given ISO 3166-1 alpha-2 code
// (case-insensitive), or an empty string when the code is unknown.
func CountryName(alpha2 string) string {
	return countryNames[strings.ToUpper(alpha2)]
}
