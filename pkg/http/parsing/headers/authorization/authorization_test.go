package authorization

import (
	"errors"
	"testing"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/google/go-cmp/cmp"
)

func TestParseAuthorization(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       []byte
		expected    *motmedelHttpTypes.Authorization
		expectedErr error
	}{
		{
			name:        "empty authorization (nil)",
			input:       nil,
			expected:    nil,
			expectedErr: motmedelErrors.ErrSyntaxError,
		},
		{
			name:        "empty authorization (empty slice)",
			input:       []byte{},
			expected:    nil,
			expectedErr: motmedelErrors.ErrSyntaxError,
		},
		{
			name:  "scheme only",
			input: []byte("Basic"),
			expected: &motmedelHttpTypes.Authorization{
				Scheme: "Basic",
			},
			expectedErr: nil,
		},
		{
			name:  "scheme with token68",
			input: []byte("Bearer abc123=="),
			expected: &motmedelHttpTypes.Authorization{
				Scheme:  "Bearer",
				Token68: "abc123==",
			},
			expectedErr: nil,
		},
		{
			name:  "scheme with single token parameter (key lowercased)",
			input: []byte("Digest Realm=foo"),
			expected: &motmedelHttpTypes.Authorization{
				Scheme: "Digest",
				Params: map[string]string{
					"realm": "foo",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "scheme with quoted parameter",
			input: []byte(`Digest realm="hello world"`),
			expected: &motmedelHttpTypes.Authorization{
				Scheme: "Digest",
				Params: map[string]string{
					"realm": "hello world",
				},
			},
			expectedErr: nil,
		},
		{
			name:  "scheme with multiple parameters and whitespace around equals and commas",
			input: []byte(`Digest realm="hello world", nonce=abc123 , opaque = "xyz"`),
			expected: &motmedelHttpTypes.Authorization{
				Scheme: "Digest",
				Params: map[string]string{
					"realm":  "hello world",
					"nonce":  "abc123",
					"opaque": "xyz",
				},
			},
			expectedErr: nil,
		},
		{
			name:        "duplicate parameter -> semantic error",
			input:       []byte("Digest a=b, a=c"),
			expected:    nil,
			expectedErr: motmedelErrors.ErrSemanticError,
		},
		{
			name:        "invalid quoted parameter value -> semantic error",
			input:       []byte(`Digest a="foo\q"`),
			expected:    nil,
			expectedErr: motmedelErrors.ErrSemanticError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			authorization, err := Parse(testCase.input)
			if !errors.Is(err, testCase.expectedErr) {
				t.Fatalf("expected error: %v, got: %v", testCase.expectedErr, err)
			}

			expected := testCase.expected

			if expected == nil && authorization != nil {
				t.Fatalf("expected nil authorization, got: %v", authorization)
			}

			if diff := cmp.Diff(expected, authorization); diff != "" {
				t.Fatalf("authorization mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
