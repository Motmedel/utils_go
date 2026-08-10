package cmp

import (
	"strings"
	"testing"
)

type diffInner struct {
	Value string
}

type diffOuter struct {
	Name       string
	Inner      *diffInner
	Tags       []string
	Attributes map[string]int
}

type diffEqualMethod struct {
	Id      int
	Ignored string
}

func (d diffEqualMethod) Equal(other diffEqualMethod) bool {
	return d.Id == other.Id
}

type diffNode struct {
	Value int
	Next  *diffNode
}

type diffUnexported struct {
	Exported string
	hidden   int //nolint:unused // exercises the unexported-field panic
}

func TestDiff(t *testing.T) {
	t.Parallel()

	selfReferentialNode := func(value int) *diffNode {
		node := &diffNode{Value: value}
		node.Next = node
		return node
	}

	testCases := []struct {
		name              string
		expected          any
		got               any
		options           []Option
		expectedSubstring string
	}{
		{
			name:     "equal strings",
			expected: "a",
			got:      "a",
		},
		{
			name:              "differing strings",
			expected:          "a",
			got:               "b",
			expectedSubstring: "(root):\n\t-: \"a\"\n\t+: \"b\"",
		},
		{
			name:     "equal structs with pointer fields",
			expected: &diffOuter{Name: "n", Inner: &diffInner{Value: "v"}, Tags: []string{"t"}},
			got:      &diffOuter{Name: "n", Inner: &diffInner{Value: "v"}, Tags: []string{"t"}},
		},
		{
			name:              "differing nested field",
			expected:          &diffOuter{Inner: &diffInner{Value: "a"}},
			got:               &diffOuter{Inner: &diffInner{Value: "b"}},
			expectedSubstring: "Inner.Value:\n\t-: \"a\"\n\t+: \"b\"",
		},
		{
			name:              "nil versus non-nil pointer field",
			expected:          &diffOuter{},
			got:               &diffOuter{Inner: &diffInner{Value: "v"}},
			expectedSubstring: "Inner:\n\t-: (*cmp.diffInner)(nil)\n\t+: &cmp.diffInner{Value: \"v\"}",
		},
		{
			name:              "type mismatch",
			expected:          1,
			got:               "1",
			expectedSubstring: "(root):\n\t-: 1\n\t+: \"1\"",
		},
		{
			name:              "differing slice element",
			expected:          &diffOuter{Tags: []string{"a", "b"}},
			got:               &diffOuter{Tags: []string{"a", "c"}},
			expectedSubstring: "Tags[1]:\n\t-: \"b\"\n\t+: \"c\"",
		},
		{
			name:              "slice length mismatch",
			expected:          &diffOuter{Tags: []string{"a"}},
			got:               &diffOuter{Tags: []string{"a", "b"}},
			expectedSubstring: "Tags:\n\t-: []string{\"a\"}\n\t+: []string{\"a\", \"b\"}",
		},
		{
			name:              "nil versus empty slice",
			expected:          []string(nil),
			got:               []string{},
			expectedSubstring: "(root):\n\t-: []string(nil)\n\t+: []string{}",
		},
		{
			name:     "nil versus empty slice with EquateEmpty",
			expected: []string(nil),
			got:      []string{},
			options:  []Option{EquateEmpty()},
		},
		{
			name:     "nil versus empty map with EquateEmpty",
			expected: map[string]int(nil),
			got:      map[string]int{},
			options:  []Option{EquateEmpty()},
		},
		{
			name:              "differing map value",
			expected:          &diffOuter{Attributes: map[string]int{"a": 1, "b": 2}},
			got:               &diffOuter{Attributes: map[string]int{"a": 1, "b": 3}},
			expectedSubstring: "Attributes[\"b\"]:\n\t-: 2\n\t+: 3",
		},
		{
			name:              "map key mismatch",
			expected:          map[string]int{"a": 1},
			got:               map[string]int{"b": 1},
			expectedSubstring: "[\"a\"]:\n\t-: 1\n\t+: <missing>",
		},
		{
			name:     "ignored fields",
			expected: &diffOuter{Name: "a", Inner: &diffInner{Value: "x"}},
			got:      &diffOuter{Name: "b", Inner: &diffInner{Value: "y"}},
			options:  []Option{IgnoreFields(diffOuter{}, "Name"), IgnoreFields(diffOuter{}, "Inner")},
		},
		{
			name:     "comparer",
			expected: diffInner{Value: "a"},
			got:      diffInner{Value: "b"},
			options: []Option{
				Comparer(func(x, y diffInner) bool { return true }),
			},
		},
		{
			name:     "equal method reporting equal",
			expected: &diffEqualMethod{Id: 1, Ignored: "a"},
			got:      &diffEqualMethod{Id: 1, Ignored: "b"},
		},
		{
			name:              "equal method reporting unequal",
			expected:          &diffEqualMethod{Id: 1, Ignored: "same"},
			got:               &diffEqualMethod{Id: 2, Ignored: "same"},
			expectedSubstring: "(root):",
		},
		{
			name:     "equal self-referential structures",
			expected: selfReferentialNode(1),
			got:      selfReferentialNode(1),
		},
		{
			name:              "differing self-referential structures",
			expected:          selfReferentialNode(1),
			got:               selfReferentialNode(2),
			expectedSubstring: "Value:\n\t-: 1\n\t+: 2",
		},
		{
			name:     "nil interfaces",
			expected: nil,
			got:      nil,
		},
		{
			name:              "nil versus non-nil interface",
			expected:          nil,
			got:               "a",
			expectedSubstring: "(root):\n\t-: <nil>\n\t+: \"a\"",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			diff := Diff(testCase.expected, testCase.got, testCase.options...)

			if testCase.expectedSubstring == "" {
				if diff != "" {
					t.Errorf("expected no diff, got:\n%s", diff)
				}
				return
			}

			if diff == "" {
				t.Errorf("expected a diff containing %q, got no diff", testCase.expectedSubstring)
			} else if !strings.Contains(diff, testCase.expectedSubstring) {
				t.Errorf("expected diff to contain %q, got:\n%s", testCase.expectedSubstring, diff)
			}
		})
	}
}

func TestDiffPanics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		function func()
	}{
		{
			name:     "IgnoreFields with non-struct type",
			function: func() { IgnoreFields(5, "Value") },
		},
		{
			name:     "IgnoreFields with unknown field",
			function: func() { IgnoreFields(diffInner{}, "Nope") },
		},
		{
			name:     "Comparer with non-function",
			function: func() { Comparer(5) },
		},
		{
			name:     "Comparer with wrong signature",
			function: func() { Comparer(func(x diffInner) bool { return true }) },
		},
		{
			name: "unexported field",
			function: func() {
				Diff(diffUnexported{Exported: "a"}, diffUnexported{Exported: "a"})
			},
		},
		{
			name: "non-nil function value",
			function: func() {
				type withFunction struct{ Function func() }
				Diff(withFunction{Function: func() {}}, withFunction{Function: func() {}})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if recover() == nil {
					t.Error("expected a panic")
				}
			}()

			testCase.function()
		})
	}
}
