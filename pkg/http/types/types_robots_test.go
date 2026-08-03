package types

import (
	"testing"
)

func TestRobotsTxtGroupString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		group    *RobotsTxtGroup
		expected string
	}{
		{
			name:     "no user agents",
			group:    &RobotsTxtGroup{Disallowed: []string{"/private"}},
			expected: "",
		},
		{
			name: "full group",
			group: &RobotsTxtGroup{
				UserAgents:   []string{"*"},
				Disallowed:   []string{"/private", "/tmp"},
				Allowed:      []string{"/public"},
				OtherRecords: [][2]string{{"Sitemap", "https://example.com/sitemap.xml"}},
			},
			expected: "User-Agent: *\n" +
				"Disallow: /private\n" +
				"Disallow: /tmp\n" +
				"Allow: /public\n" +
				"Sitemap: https://example.com/sitemap.xml",
		},
		{
			name: "empty disallow value allowed",
			group: &RobotsTxtGroup{
				UserAgents: []string{"*"},
				Disallowed: []string{""},
			},
			expected: "User-Agent: *\nDisallow: ",
		},
		{
			name: "empty allow value skipped",
			group: &RobotsTxtGroup{
				UserAgents: []string{"Googlebot"},
				Allowed:    []string{""},
			},
			expected: "User-Agent: Googlebot",
		},
		{
			name: "empty other record skipped",
			group: &RobotsTxtGroup{
				UserAgents:   []string{"*"},
				OtherRecords: [][2]string{{"Sitemap", ""}},
			},
			expected: "User-Agent: *",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.group.String(); got != testCase.expected {
				t.Errorf("String() = %q, want %q", got, testCase.expected)
			}
		})
	}
}

func TestRobotsTxtString(t *testing.T) {
	t.Parallel()

	group1 := &RobotsTxtGroup{UserAgents: []string{"*"}, Disallowed: []string{"/a"}}
	group2 := &RobotsTxtGroup{UserAgents: []string{"Googlebot"}, Disallowed: []string{"/b"}}
	emptyGroup := &RobotsTxtGroup{Disallowed: []string{"/c"}}

	testCases := []struct {
		name      string
		robotsTxt *RobotsTxt
		expected  string
	}{
		{
			name:      "empty",
			robotsTxt: &RobotsTxt{},
			expected:  "",
		},
		{
			name:      "two groups with nil and empty skipped",
			robotsTxt: &RobotsTxt{Groups: []*RobotsTxtGroup{group1, nil, emptyGroup, group2}},
			expected:  "User-Agent: *\nDisallow: /a\n\nUser-Agent: Googlebot\nDisallow: /b",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.robotsTxt.String(); got != testCase.expected {
				t.Errorf("String() = %q, want %q", got, testCase.expected)
			}
		})
	}
}
