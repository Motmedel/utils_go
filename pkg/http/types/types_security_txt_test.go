package types

import (
	"testing"
	"time"
)

func TestSecurityTxtString(t *testing.T) {
	t.Parallel()

	expires := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)

	testCases := []struct {
		name        string
		securityTxt *SecurityTxt
		expected    string
	}{
		{
			name:        "nil",
			securityTxt: nil,
			expected:    "",
		},
		{
			// RFC 9116 requires a contact; without one there is nothing to say to a reporter.
			name:        "without contacts",
			securityTxt: &SecurityTxt{Expires: expires, PreferredLanguages: []string{"en"}},
			expected:    "",
		},
		{
			name:        "a contact alone",
			securityTxt: &SecurityTxt{Contacts: []string{"mailto:security@example.com"}},
			expected:    "Contact: mailto:security@example.com\n",
		},
		{
			name: "contacts in the order they are preferred",
			securityTxt: &SecurityTxt{
				Contacts: []string{"mailto:security@example.com", "https://example.com/vulnerability"},
				Expires:  expires,
			},
			expected: "Contact: mailto:security@example.com\n" +
				"Contact: https://example.com/vulnerability\n" +
				"Expires: 2030-01-02T03:04:05Z\n",
		},
		{
			// Preferred-Languages is one line listing the languages, unlike the repeated fields.
			name: "everything",
			securityTxt: &SecurityTxt{
				Contacts:           []string{"mailto:security@example.com"},
				Expires:            expires,
				Encryption:         []string{"https://example.com/pgp-key.txt"},
				Acknowledgments:    []string{"https://example.com/thanks"},
				PreferredLanguages: []string{"sv", "en"},
				Canonical:          []string{"https://example.com/.well-known/security.txt"},
				Policy:             []string{"https://example.com/disclosure"},
				Hiring:             []string{"https://example.com/jobs"},
				Csaf:               []string{"https://example.com/.well-known/csaf/provider-metadata.json"},
			},
			expected: "Contact: mailto:security@example.com\n" +
				"Expires: 2030-01-02T03:04:05Z\n" +
				"Encryption: https://example.com/pgp-key.txt\n" +
				"Acknowledgments: https://example.com/thanks\n" +
				"Preferred-Languages: sv, en\n" +
				"Canonical: https://example.com/.well-known/security.txt\n" +
				"Policy: https://example.com/disclosure\n" +
				"Hiring: https://example.com/jobs\n" +
				"CSAF: https://example.com/.well-known/csaf/provider-metadata.json\n",
		},
		{
			// The instant is stated in UTC, as RFC 9116 wants it, whatever it was given in.
			name: "expires in another zone",
			securityTxt: &SecurityTxt{
				Contacts: []string{"mailto:security@example.com"},
				Expires:  expires.In(time.FixedZone("CET", 3600)),
			},
			expected: "Contact: mailto:security@example.com\nExpires: 2030-01-02T03:04:05Z\n",
		},
		{
			name: "empty values are left out",
			securityTxt: &SecurityTxt{
				Contacts:   []string{"mailto:security@example.com", "  "},
				Encryption: []string{""},
			},
			expected: "Contact: mailto:security@example.com\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.securityTxt.String(); got != testCase.expected {
				t.Errorf("string:\ngot:\n%s\nwant:\n%s", got, testCase.expected)
			}
		})
	}
}
