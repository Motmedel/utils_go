package cmp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type simpleError struct {
	Msg string
}

func (e *simpleError) Error() string {
	return e.Msg
}

type otherError struct {
	Code int
}

func (e *otherError) Error() string {
	return fmt.Sprintf("other: %d", e.Code)
}

type detailError struct {
	Msg    string
	Detail string
}

func (e *detailError) Error() string {
	return e.Msg
}

// TestCompareErrMatching exercises the inputs for which CompareErr must not
// report any failure. The invocations run against the subtest's own
// *testing.T, so an unexpected Fatalf/Errorf inside CompareErr fails the test.
func TestCompareErrMatching(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  error
		want error
		opts []Option
	}{
		{
			name: "both nil",
			got:  nil,
			want: nil,
		},
		{
			name: "identical typed errors",
			got:  &simpleError{Msg: "boom"},
			want: &simpleError{Msg: "boom"},
		},
		{
			name: "wrapped got matches unwrapped want",
			got:  fmt.Errorf("context: %w", &simpleError{Msg: "boom"}),
			want: &simpleError{Msg: "boom"},
		},
		{
			name: "comparer option ignores differing field",
			got:  &detailError{Msg: "same", Detail: "actual"},
			want: &detailError{Msg: "same", Detail: "expected"},
			opts: []Option{
				Comparer(func(x, y *detailError) bool {
					return x.Msg == y.Msg
				}),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			CompareErr(t, testCase.got, testCase.want, testCase.opts...)
		})
	}
}

// TestCompareErrSubprocess is re-executed as a child process by
// TestCompareErrFailing. It performs a single CompareErr invocation that is
// expected to fail, selected by the CMP_SCENARIO environment variable.
func TestCompareErrSubprocess(t *testing.T) {
	t.Parallel()

	switch os.Getenv("CMP_SCENARIO") {
	case "want-nil-got-set":
		CompareErr(t, &simpleError{Msg: "boom"}, nil)
	case "want-set-got-nil":
		CompareErr(t, nil, &simpleError{Msg: "boom"})
	case "type-mismatch":
		CompareErr(t, &otherError{Code: 7}, &simpleError{Msg: "boom"})
	case "value-mismatch":
		CompareErr(t, &simpleError{Msg: "actual"}, &simpleError{Msg: "expected"})
	default:
		t.Skip("TestCompareErrSubprocess is only meaningful when CMP_SCENARIO is set")
	}
}

// runCompareErrScenario re-executes the test binary so that a failing
// CompareErr invocation can be observed in isolation. It returns the combined
// subprocess output and whether the subprocess reported a failure.
func runCompareErrScenario(t *testing.T, scenario string) (string, bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	// The command is this test binary re-executing a single known test; the
	// only variable is os.Args[0], which is not attacker-controlled.
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCompareErrSubprocess$", "-test.v") //nolint:gosec // re-executing the test binary
	cmd.Env = append(os.Environ(), "CMP_SCENARIO="+scenario)

	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), false
	}

	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		return string(output), true
	}

	t.Fatalf("failed to run CompareErr subprocess for scenario %q: %v\noutput:\n%s", scenario, err, output)
	return string(output), false
}

// TestCompareErrFailing exercises the inputs for which CompareErr must report a
// failure, asserting both the failure and the formatted message it produces.
func TestCompareErrFailing(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		scenario   string
		wantSubstr string
	}{
		{
			name:       "want nil but got a non-nil error",
			scenario:   "want-nil-got-set",
			wantSubstr: "expected no error, got *cmp.simpleError: boom",
		},
		{
			name:       "want an error but got nil",
			scenario:   "want-set-got-nil",
			wantSubstr: "expected *cmp.simpleError error, got nil",
		},
		{
			name:       "got error not assignable to want type",
			scenario:   "type-mismatch",
			wantSubstr: "expected error assignable to *cmp.simpleError, got *cmp.otherError",
		},
		{
			name:       "matching type but differing value",
			scenario:   "value-mismatch",
			wantSubstr: "error mismatch (-expected +got):",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output, failed := runCompareErrScenario(t, testCase.scenario)
			if !failed {
				t.Fatalf(
					"expected CompareErr to fail for scenario %q, but the subprocess succeeded; output:\n%s",
					testCase.scenario, output,
				)
			}

			if !strings.Contains(output, testCase.wantSubstr) {
				t.Errorf(
					"CompareErr failure output for scenario %q missing %q; got:\n%s",
					testCase.scenario, testCase.wantSubstr, output,
				)
			}
		})
	}
}

type comparableUnexported struct {
	value int
}

type equateComparableHolder struct {
	Inner comparableUnexported
}

func TestEquateComparable(t *testing.T) {
	t.Parallel()

	equal := equateComparableHolder{Inner: comparableUnexported{value: 1}}
	alsoEqual := equateComparableHolder{Inner: comparableUnexported{value: 1}}
	different := equateComparableHolder{Inner: comparableUnexported{value: 2}}

	if diff := Diff(equal, alsoEqual, EquateComparable(comparableUnexported{})); diff != "" {
		t.Errorf("expected no diff, got:\n%s", diff)
	}

	if diff := Diff(equal, different, EquateComparable(comparableUnexported{})); diff == "" {
		t.Errorf("expected a diff")
	}
}

func TestEquateComparableRejectsUncomparable(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Errorf("expected panic")
		}
	}()

	EquateComparable([]int{})
}

func TestDiffPanicsOnUnexportedWithoutOption(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Errorf("expected panic")
		}
	}()

	_ = Diff(
		equateComparableHolder{Inner: comparableUnexported{value: 1}},
		equateComparableHolder{Inner: comparableUnexported{value: 2}},
	)
}
