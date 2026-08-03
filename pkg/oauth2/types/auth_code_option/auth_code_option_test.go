package auth_code_option

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		key   string
		value string
	}{
		{name: "typical", key: "access_type", value: "offline"},
		{name: "empty key", key: "", value: "value"},
		{name: "empty value", key: "prompt", value: ""},
		{name: "both empty", key: "", value: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			option := New(testCase.key, testCase.value)
			if option.Key != testCase.key {
				t.Errorf("Key = %q, want %q", option.Key, testCase.key)
			}
			if option.Value != testCase.value {
				t.Errorf("Value = %q, want %q", option.Value, testCase.value)
			}
		})
	}
}
