package firewall_verdict

import "testing"

func TestVerdictConstantValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		verdict Verdict
		want    int
	}{
		{name: "Accept", verdict: Accept, want: 0},
		{name: "Drop", verdict: Drop, want: 1},
		{name: "Reject", verdict: Reject, want: 2},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if int(testCase.verdict) != testCase.want {
				t.Errorf("%s = %d, want %d", testCase.name, int(testCase.verdict), testCase.want)
			}
		})
	}
}

func TestVerdictZeroValueIsAccept(t *testing.T) {
	t.Parallel()

	var zero Verdict
	if zero != Accept {
		t.Errorf("zero value = %d, want Accept (%d)", int(zero), int(Accept))
	}
}

func TestVerdictConstantsAreDistinct(t *testing.T) {
	t.Parallel()

	if Accept == Drop || Accept == Reject || Drop == Reject {
		t.Errorf("Verdict constants are not distinct: Accept=%d Drop=%d Reject=%d",
			int(Accept), int(Drop), int(Reject))
	}
}
