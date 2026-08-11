package integrity_policy

import (
	"encoding/json/v2"
	"testing"
)

func TestIntegrityViolationReportBodyMessage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		body     *IntegrityViolationReportBody
		expected string
	}{
		{
			name: "script destination",
			body: &IntegrityViolationReportBody{
				Destination: "script",
				BlockedUrl:  "https://example.com/app.js",
			},
			expected: "The page's settings blocked a script at https://example.com/app.js from being loaded because it is missing integrity metadata.",
		},
		{
			name:     "empty fields",
			body:     &IntegrityViolationReportBody{},
			expected: "The page's settings blocked a  at  from being loaded because it is missing integrity metadata.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got := testCase.body.Message(); got != testCase.expected {
				t.Errorf("Message() = %q, want %q", got, testCase.expected)
			}
		})
	}
}

func TestIntegrityViolationReportBodyJSON(t *testing.T) {
	t.Parallel()

	body := &IntegrityViolationReportBody{
		DocumentUrl: "https://example.com/",
		BlockedUrl:  "https://example.com/app.js",
		Destination: "script",
		ReportOnly:  true,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var roundTrip map[string]any
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if roundTrip["documentURL"] != "https://example.com/" {
		t.Errorf("documentURL = %v, want %v", roundTrip["documentURL"], body.DocumentUrl)
	}
	if roundTrip["blockedURL"] != "https://example.com/app.js" {
		t.Errorf("blockedURL = %v, want %v", roundTrip["blockedURL"], body.BlockedUrl)
	}
	if roundTrip["destination"] != "script" {
		t.Errorf("destination = %v, want %v", roundTrip["destination"], body.Destination)
	}
	if roundTrip["reportOnly"] != true {
		t.Errorf("reportOnly = %v, want true", roundTrip["reportOnly"])
	}
}
