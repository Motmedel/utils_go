package tree

import (
	"log/slog"
	"reflect"
	"testing"
)

// attrsToMap resolves a rendered attribute slice into a comparable map.
func attrsToMap(attrs []slog.Attr) map[string]any {
	m := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		value := attr.Value.Resolve()
		if value.Kind() == slog.KindGroup {
			m[attr.Key] = attrsToMap(value.Group())
		} else {
			m[attr.Key] = value.Any()
		}
	}
	return m
}

func TestGroupMergeAndRender(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		build func(group *Group)
		want  map[string]any
	}{
		{
			name: "single scalar",
			build: func(group *Group) {
				group.Merge(nil, slog.String("a", "1"))
			},
			want: map[string]any{"a": "1"},
		},
		{
			name: "multiple scalars",
			build: func(group *Group) {
				group.Merge(nil, slog.String("a", "1"), slog.String("b", "2"))
			},
			want: map[string]any{"a": "1", "b": "2"},
		},
		{
			name: "scalar override",
			build: func(group *Group) {
				group.Merge(nil, slog.String("a", "1"))
				group.Merge(nil, slog.String("a", "2"))
			},
			want: map[string]any{"a": "2"},
		},
		{
			name: "group creation",
			build: func(group *Group) {
				group.Merge(nil, slog.Group("g", slog.String("x", "1")))
			},
			want: map[string]any{"g": map[string]any{"x": "1"}},
		},
		{
			name: "merge into existing group",
			build: func(group *Group) {
				group.Merge(nil, slog.Group("g", slog.String("x", "1")))
				group.Merge(nil, slog.Group("g", slog.String("y", "2")))
			},
			want: map[string]any{"g": map[string]any{"x": "1", "y": "2"}},
		},
		{
			name: "scalar overrides group",
			build: func(group *Group) {
				group.Merge(nil, slog.Group("g", slog.String("x", "1")))
				group.Merge(nil, slog.String("g", "5"))
			},
			want: map[string]any{"g": "5"},
		},
		{
			name: "group overrides scalar",
			build: func(group *Group) {
				group.Merge(nil, slog.String("g", "5"))
				group.Merge(nil, slog.Group("g", slog.String("x", "1")))
			},
			want: map[string]any{"g": map[string]any{"x": "1"}},
		},
		{
			name: "nested via stack",
			build: func(group *Group) {
				group.Merge(nil, slog.Group("a"))
				group.Merge([]string{"a"}, slog.String("x", "1"))
			},
			want: map[string]any{"a": map[string]any{"x": "1"}},
		},
		{
			name: "deep nested via stack",
			build: func(group *Group) {
				group.Merge(nil, slog.Group("a"))
				group.Merge([]string{"a"}, slog.Group("b"))
				group.Merge([]string{"a", "b"}, slog.String("y", "2"))
			},
			want: map[string]any{"a": map[string]any{"b": map[string]any{"y": "2"}}},
		},
		{
			name: "int value",
			build: func(group *Group) {
				group.Merge(nil, slog.Int("n", 1))
			},
			want: map[string]any{"n": int64(1)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			group := new(Group)
			testCase.build(group)

			got := attrsToMap(group.Render())
			if !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("got %#v, want %#v", got, testCase.want)
			}
		})
	}
}

func TestGroupCloneIsDeep(t *testing.T) {
	t.Parallel()

	original := new(Group)
	original.Merge(nil, slog.Group("g", slog.String("x", "1")))

	clone := original.Clone()
	// Mutate the clone; the original must remain unchanged.
	clone.Merge([]string{"g"}, slog.String("y", "2"))
	clone.Merge(nil, slog.String("top", "new"))

	gotOriginal := attrsToMap(original.Render())
	wantOriginal := map[string]any{"g": map[string]any{"x": "1"}}
	if !reflect.DeepEqual(gotOriginal, wantOriginal) {
		t.Fatalf("original mutated: got %#v, want %#v", gotOriginal, wantOriginal)
	}

	gotClone := attrsToMap(clone.Render())
	wantClone := map[string]any{"g": map[string]any{"x": "1", "y": "2"}, "top": "new"}
	if !reflect.DeepEqual(gotClone, wantClone) {
		t.Fatalf("clone: got %#v, want %#v", gotClone, wantClone)
	}
}

func TestValueLogValue(t *testing.T) {
	t.Parallel()

	scalar := &Value{Name: "a", SlogValue: slog.StringValue("x")}
	if got := scalar.LogValue(); got.Kind() != slog.KindString || got.String() != "x" {
		t.Fatalf("scalar LogValue = %v, want string x", got)
	}

	group := &Group{Values: []*Value{{Name: "x", SlogValue: slog.StringValue("1")}}}
	holder := &Value{Name: "g", Group: group}
	got := holder.LogValue()
	if got.Kind() != slog.KindGroup {
		t.Fatalf("group holder LogValue kind = %v, want group", got.Kind())
	}
	if want := (map[string]any{"x": "1"}); !reflect.DeepEqual(attrsToMap(got.Group()), want) {
		t.Fatalf("group holder LogValue = %#v, want %#v", attrsToMap(got.Group()), want)
	}
}

func TestGroupLogValueMatchesRender(t *testing.T) {
	t.Parallel()

	group := new(Group)
	group.Merge(nil, slog.String("a", "1"), slog.Group("g", slog.String("x", "2")))

	logValue := group.LogValue()
	if logValue.Kind() != slog.KindGroup {
		t.Fatalf("Group.LogValue kind = %v, want group", logValue.Kind())
	}
	if !reflect.DeepEqual(attrsToMap(logValue.Group()), attrsToMap(group.Render())) {
		t.Fatal("Group.LogValue does not match Render")
	}
}

func TestMergeInvalidStackPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid stack, got none")
		}
	}()

	group := new(Group)
	group.Merge([]string{"missing"}, slog.String("a", "1"))
}
