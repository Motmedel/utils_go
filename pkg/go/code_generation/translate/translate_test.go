package translate

import "testing"

func TestMap(t *testing.T) {
	t.Parallel()

	out, err := Map(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error for empty map: %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty string for empty map, got: %q", out)
	}

	m := map[string]any{
		"name":     "Alice",                        // const
		"enabled":  true,                           // const
		"pi":       3.14,                           // const (float64)
		"count":    42,                             // var (int)
		"nums":     []int{1, 2},                    // var (slice)
		"data":     map[string]int{"a": 1, "b": 2}, // var (map, order not guaranteed)
		"9bad key": "x",                            // const, ident becomes _9bad_key
		"package":  "p",                            // const, keyword becomes package_
	}

	out, err = Map(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Build exact expected output. Keys should be sorted within const and var blocks.
	expected1 := "const (\n" +
		"_9bad_key = \"x\"\n" +
		"enabled = true\n" +
		"name = \"Alice\"\n" +
		"package_ = \"p\"\n" +
		"pi = 3.14\n" +
		")\n\n" +
		"var (\n" +
		"count = 42\n" +
		"data = map[string]int{\"a\": 1, \"b\": 2}\n" +
		"nums = []int{1, 2}\n" +
		")"

	// Same but with the other possible map key order inside the map literal
	expected2 := "const (\n" +
		"_9bad_key = \"x\"\n" +
		"enabled = true\n" +
		"name = \"Alice\"\n" +
		"package_ = \"p\"\n" +
		"pi = 3.14\n" +
		")\n\n" +
		"var (\n" +
		"count = 42\n" +
		"data = map[string]int{\"b\": 2, \"a\": 1}\n" +
		"nums = []int{1, 2}\n" +
		")"

	if out != expected1 && out != expected2 {
		t.Fatalf("unexpected output.\nGot:\n%s\n\nWant one of:\n%s\n-- OR --\n%s", out, expected1, expected2)
	}

	// Error case: non-nil function value should cause an error
	_, err = Map(map[string]any{"bad": func() {}})
	if err == nil {
		t.Errorf("expected error for non-nil function value, got nil")
	}
}

func Test_isConst(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   any
		out  bool
	}{
		{"string", "hello", true},
		{"bool true", true, true},
		{"bool false", false, true},
		{"float64", 1.23, true},
		{"int", 7, false},
		{"map", map[string]int{"a": 1}, false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isConst(tc.in); got != tc.out {
				t.Fatalf("isConst(%T) = %v, want %v", tc.in, got, tc.out)
			}
		})
	}
}

func Test_toIdent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
		name string
	}{
		{"", "_", "empty"},
		{"hello", "hello", "plain"},
		{"with space", "with_space", "space to underscore"},
		{"dollar$sign", "dollar_sign", "non-word to underscore"},
		{"9lives", "_9lives", "leading digit"},
		{"package", "package_", "keyword"},
		{"already_ok_123", "already_ok_123", "unchanged"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := toIdent(tc.in); got != tc.want {
				t.Fatalf("toIdent(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
