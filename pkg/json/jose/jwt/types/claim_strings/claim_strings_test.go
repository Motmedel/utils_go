package claim_strings

import (
	"encoding/json"
	"errors"
	"testing"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

func TestClaimStrings_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		expected    ClaimStrings
		expectError bool
	}{
		{
			name:     "single string",
			input:    `"audience1"`,
			expected: ClaimStrings{"audience1"},
		},
		{
			name:     "array of strings",
			input:    `["audience1", "audience2", "audience3"]`,
			expected: ClaimStrings{"audience1", "audience2", "audience3"},
		},
		{
			name:     "empty array",
			input:    `[]`,
			expected: ClaimStrings{},
		},
		{
			name:     "null value",
			input:    `null`,
			expected: nil,
		},
		{
			name:        "invalid json",
			input:       `{invalid`,
			expectError: true,
		},
		{
			name:        "array with non-string",
			input:       `["valid", 123]`,
			expectError: true,
		},
		{
			name:        "number instead of string",
			input:       `123`,
			expectError: true,
		},
		{
			name:        "object instead of string",
			input:       `{"key": "value"}`,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var cs ClaimStrings
			err := json.Unmarshal([]byte(tc.input), &cs)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(cs) != len(tc.expected) {
					t.Fatalf("expected length %d, got %d", len(tc.expected), len(cs))
				}

				for i, v := range tc.expected {
					if cs[i] != v {
						t.Fatalf("expected %s at index %d, got %s", v, i, cs[i])
					}
				}
			}
		})
	}
}

func TestClaimStrings_MarshalJSON(t *testing.T) {
	t.Parallel()

	// Save and restore the global setting
	originalSetting := MarshalSingleStringAsArray

	testCases := []struct {
		name                      string
		input                     ClaimStrings
		marshalSingleStringAsArr  bool
		expected                  string
	}{
		{
			name:                     "single string as array",
			input:                    ClaimStrings{"audience1"},
			marshalSingleStringAsArr: true,
			expected:                 `["audience1"]`,
		},
		{
			name:                     "single string as string",
			input:                    ClaimStrings{"audience1"},
			marshalSingleStringAsArr: false,
			expected:                 `"audience1"`,
		},
		{
			name:                     "multiple strings always as array",
			input:                    ClaimStrings{"audience1", "audience2"},
			marshalSingleStringAsArr: false,
			expected:                 `["audience1","audience2"]`,
		},
		{
			name:                     "empty array",
			input:                    ClaimStrings{},
			marshalSingleStringAsArr: true,
			expected:                 `[]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: Not using t.Parallel() here because we modify global state
			MarshalSingleStringAsArray = tc.marshalSingleStringAsArr

			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(b) != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, string(b))
			}
		})
	}

	// Restore original setting
	MarshalSingleStringAsArray = originalSetting
}

func TestConvert(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       any
		expected    ClaimStrings
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name:     "single string",
			input:    "audience1",
			expected: ClaimStrings{"audience1"},
		},
		{
			name:     "string slice",
			input:    []string{"audience1", "audience2"},
			expected: ClaimStrings{"audience1", "audience2"},
		},
		{
			name:     "any slice with strings",
			input:    []any{"audience1", "audience2"},
			expected: ClaimStrings{"audience1", "audience2"},
		},
		{
			name:        "any slice with non-string",
			input:       []any{"audience1", 123},
			expectError: true,
		},
		{
			name:        "integer (unsupported)",
			input:       123,
			expectError: true,
			errorCheck: func(err error) bool {
				return errors.Is(err, motmedelErrors.ErrUnexpectedType)
			},
		},
		{
			name:        "nil (unsupported)",
			input:       nil,
			expectError: true,
			errorCheck: func(err error) bool {
				return errors.Is(err, motmedelErrors.ErrUnexpectedType)
			},
		},
		{
			name:        "map (unsupported)",
			input:       map[string]string{"key": "value"},
			expectError: true,
			errorCheck: func(err error) bool {
				return errors.Is(err, motmedelErrors.ErrUnexpectedType)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := Convert(tc.input)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				if tc.errorCheck != nil && !tc.errorCheck(err) {
					t.Fatalf("error check failed for error: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(result) != len(tc.expected) {
					t.Fatalf("expected length %d, got %d", len(tc.expected), len(result))
				}

				for i, v := range tc.expected {
					if result[i] != v {
						t.Fatalf("expected %s at index %d, got %s", v, i, result[i])
					}
				}
			}
		})
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input ClaimStrings
	}{
		{
			name:  "single string",
			input: ClaimStrings{"audience1"},
		},
		{
			name:  "multiple strings",
			input: ClaimStrings{"audience1", "audience2", "audience3"},
		},
		{
			name:  "empty",
			input: ClaimStrings{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var decoded ClaimStrings
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if len(decoded) != len(tc.input) {
				t.Fatalf("round trip failed: expected length %d, got %d", len(tc.input), len(decoded))
			}

			for i, v := range tc.input {
				if decoded[i] != v {
					t.Fatalf("round trip failed at index %d: expected %s, got %s", i, v, decoded[i])
				}
			}
		})
	}
}
