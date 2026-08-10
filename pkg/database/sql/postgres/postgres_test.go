package postgres

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

func TestParseTextArray(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		data        string
		expected    []string
		expectError bool
	}{
		{name: "empty", data: "{}", expected: []string{}},
		{name: "single", data: "{admin}", expected: []string{"admin"}},
		{name: "multiple", data: "{admin,user}", expected: []string{"admin", "user"}},
		{name: "quoted", data: `{"hello world"}`, expected: []string{"hello world"}},
		{name: "quoted comma", data: `{"a,b",c}`, expected: []string{"a,b", "c"}},
		{name: "escaped quote", data: `{"say \"hi\""}`, expected: []string{`say "hi"`}},
		{name: "escaped backslash", data: `{"back\\slash"}`, expected: []string{`back\slash`}},
		{name: "quoted braces", data: `{"{curly}"}`, expected: []string{"{curly}"}},
		{name: "quoted empty", data: `{""}`, expected: []string{""}},
		{name: "quoted null word", data: `{"NULL"}`, expected: []string{"NULL"}},
		{name: "mixed", data: `{plain,"quoted, too",tail}`, expected: []string{"plain", "quoted, too", "tail"}},
		{name: "dimension prefix", data: "[0:1]={a,b}", expected: []string{"a", "b"}},
		{name: "missing braces", data: "a,b", expectError: true},
		{name: "empty input", data: "", expectError: true},
		{name: "unterminated quote", data: `{"a`, expectError: true},
		{name: "trailing escape", data: `{"a\`, expectError: true},
		{name: "null element", data: "{NULL}", expectError: true},
		{name: "empty unquoted element", data: "{a,,b}", expectError: true},
		{name: "missing separator", data: `{"a"b}`, expectError: true},
		{name: "unquoted quote", data: `{a"b}`, expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			elements, err := ParseTextArray(testCase.data)
			if testCase.expectError {
				if !errors.Is(err, ErrMalformedTextArray) {
					t.Fatalf("expected malformed text array error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse text array: %v", err)
			}

			if !slices.Equal(elements, testCase.expected) {
				t.Errorf("elements: got %q, want %q", elements, testCase.expected)
			}
		})
	}
}

func TestTextArrayScanner(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		value       any
		expected    []string
		expectError bool
	}{
		{name: "bytes", value: []byte("{a,b}"), expected: []string{"a", "b"}},
		{name: "string", value: "{a}", expected: []string{"a"}},
		{name: "null", value: nil, expected: nil},
		{name: "unsupported type", value: 1, expectError: true},
		{name: "malformed", value: "a", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var target []string
			err := TextArrayScanner{Target: &target}.Scan(testCase.value)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("scan: %v", err)
			}

			if !slices.Equal(target, testCase.expected) {
				t.Errorf("target: got %q, want %q", target, testCase.expected)
			}
		})
	}
}

type testSqlStateError struct {
	code string
}

func (e *testSqlStateError) Error() string {
	return "test error"
}

func (e *testSqlStateError) SQLState() string {
	return e.code
}

func TestSqlState(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		err           error
		expectedCode  string
		expectedFound bool
	}{
		{
			name:          "direct",
			err:           &testSqlStateError{code: SqlStateUniqueViolation},
			expectedCode:  SqlStateUniqueViolation,
			expectedFound: true,
		},
		{
			name:          "wrapped",
			err:           fmt.Errorf("scan: %w", &testSqlStateError{code: "23503"}),
			expectedCode:  "23503",
			expectedFound: true,
		},
		{name: "plain error", err: ErrMalformedTextArray},
		{name: "nil", err: nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			code, found := SqlState(testCase.err)
			if code != testCase.expectedCode || found != testCase.expectedFound {
				t.Errorf(
					"SqlState: got (%q, %t), want (%q, %t)",
					code, found, testCase.expectedCode, testCase.expectedFound,
				)
			}
		})
	}
}
