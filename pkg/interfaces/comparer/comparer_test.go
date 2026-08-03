package comparer

import (
	"errors"
	"testing"
)

var errSentinel = errors.New("comparer failure")

// mockComparer is a minimal implementation of the Comparer interface.
type mockComparer struct {
	result bool
	err    error
}

func (m *mockComparer) Compare(_ int) (bool, error) {
	return m.result, m.err
}

// Compile-time assertions that the mock, the Function adapter, and AnyEqualComparer satisfy the interface.
var (
	_ Comparer[int] = (*mockComparer)(nil)
	_ Comparer[int] = Function[int](nil)
	_ Comparer[int] = AnyEqualComparer[int]{}
	_ Comparer[int] = (*AnyEqualComparer[int])(nil)
)

func TestMockThroughInterface(t *testing.T) {
	t.Parallel()

	var comparer Comparer[int] = &mockComparer{result: true}

	got, err := comparer.Compare(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Fatal("expected true")
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	comparer := New(func(input string) (bool, error) {
		return input == "match", errSentinel
	})

	got, err := comparer.Compare("match")
	if !got {
		t.Fatal("expected true")
	}
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestAnyEqualComparer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		values   []int
		input    int
		expected bool
	}{
		{name: "present first", values: []int{1, 2, 3}, input: 1, expected: true},
		{name: "present middle", values: []int{1, 2, 3}, input: 2, expected: true},
		{name: "present last", values: []int{1, 2, 3}, input: 3, expected: true},
		{name: "absent", values: []int{1, 2, 3}, input: 4, expected: false},
		{name: "empty values", values: nil, input: 1, expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			comparer := NewEqualComparer(testCase.values...)
			got, err := comparer.Compare(testCase.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != testCase.expected {
				t.Fatalf("expected %v, got %v", testCase.expected, got)
			}
		})
	}
}

func TestNewEqualComparerStoresValues(t *testing.T) {
	t.Parallel()

	comparer := NewEqualComparer("a", "b")
	if comparer == nil {
		t.Fatal("expected non-nil comparer")
	}
	if len(comparer.Values) != 2 || comparer.Values[0] != "a" || comparer.Values[1] != "b" {
		t.Fatalf("unexpected stored values: %v", comparer.Values)
	}
}
