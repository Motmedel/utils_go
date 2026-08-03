package body_setting

import "testing"

func TestSettingConstantValues(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		setting Setting
		want    int
	}{
		{name: "Required", setting: Required, want: 0},
		{name: "Optional", setting: Optional, want: 1},
		{name: "Forbidden", setting: Forbidden, want: 2},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if int(testCase.setting) != testCase.want {
				t.Errorf("%s = %d, want %d", testCase.name, int(testCase.setting), testCase.want)
			}
		})
	}
}

func TestSettingZeroValueIsRequired(t *testing.T) {
	t.Parallel()

	var zero Setting
	if zero != Required {
		t.Errorf("zero value = %d, want Required (%d)", int(zero), int(Required))
	}
}

func TestSettingConstantsAreDistinct(t *testing.T) {
	t.Parallel()

	if Required == Optional || Required == Forbidden || Optional == Forbidden {
		t.Errorf("Setting constants are not distinct: Required=%d Optional=%d Forbidden=%d",
			int(Required), int(Optional), int(Forbidden))
	}
}
