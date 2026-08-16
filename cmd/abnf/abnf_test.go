package main

import (
	"bytes"
	"encoding/json/v2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testGrammar = "; why\r\nroot = ( \"a\" / \"b\" )  1*1DIGIT\r\n"

const testMinifiedGrammar = "root=(\"a\"/\"b\") DIGIT\r\n"

const testSimplifiedGrammar = "root=(\"a\"/\"b\") DIGIT\r\n"

// runMain invokes run() with a controlled argv, feeding it the given standard
// input and capturing anything written to standard output. os.Args, os.Stdin
// and os.Stdout are saved and restored so the call can be repeated. These
// tests must not run in parallel because they mutate this process-wide state.
func runMain(t *testing.T, stdin string, arguments ...string) (string, int) {
	t.Helper()

	origArgs, origStdin, origStdout := os.Args, os.Stdin, os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Args, os.Stdin, os.Stdout, os.Stderr = origArgs, origStdin, origStdout, origStderr
	}()

	os.Args = append([]string{"abnf"}, arguments...)

	// Usage and diagnostics go to standard error; the tests read exit codes
	// rather than that text, so discard it.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("failed to open the null device: %v", err)
	}
	defer func() { _ = devNull.Close() }()
	os.Stderr = devNull

	stdinPath := filepath.Join(t.TempDir(), "stdin")
	if err := os.WriteFile(stdinPath, []byte(stdin), 0o600); err != nil {
		t.Fatalf("failed to write the stdin file: %v", err)
	}
	stdinFile, openErr := os.Open(stdinPath)
	if openErr != nil {
		t.Fatalf("failed to open the stdin file: %v", openErr)
	}
	defer func() { _ = stdinFile.Close() }()
	os.Stdin = stdinFile

	readEnd, writeEnd, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("failed to create stdout pipe: %v", pipeErr)
	}
	os.Stdout = writeEnd

	captured := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, readEnd)
		captured <- buf.String()
	}()

	code, runErr := run()
	if runErr != nil {
		t.Fatalf("unexpected error: %v", runErr)
	}

	_ = writeEnd.Close()
	stdout := <-captured
	_ = readEnd.Close()

	return stdout, code
}

// writeGrammar writes a grammar definition to a temporary file and returns
// its path.
func writeGrammar(t *testing.T, grammar string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "grammar.abnf")
	if err := os.WriteFile(path, []byte(grammar), 0o600); err != nil {
		t.Fatalf("failed to write the grammar file: %v", err)
	}

	return path
}

func TestMinify(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	testCases := []struct {
		name      string
		arguments []string
		stdin     string
		expected  string
	}{
		{name: "stdin", arguments: []string{"minify"}, stdin: testGrammar, expected: testMinifiedGrammar},
		{name: "simplify", arguments: []string{"minify", "--simplify"}, stdin: testGrammar, expected: testSimplifiedGrammar},
	}

	for _, testCase := range testCases { //nolint:paralleltest // shares process-global state through run()
		t.Run(testCase.name, func(t *testing.T) {
			stdout, code := runMain(t, testCase.stdin, testCase.arguments...)
			if code != exitClean {
				t.Fatalf("expected exit %d, got %d", exitClean, code)
			}
			if stdout != testCase.expected {
				t.Fatalf("expected %q, got %q", testCase.expected, stdout)
			}
		})
	}
}

func TestMinifyPath(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	stdout, code := runMain(t, "", "minify", writeGrammar(t, testGrammar))
	if code != exitClean {
		t.Fatalf("expected exit %d, got %d", exitClean, code)
	}
	if stdout != testMinifiedGrammar {
		t.Fatalf("expected %q, got %q", testMinifiedGrammar, stdout)
	}
}

func TestMinifyOutput(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	outputPath := filepath.Join(t.TempDir(), "minified.abnf")

	stdout, code := runMain(t, "", "minify", "--output", outputPath, writeGrammar(t, testGrammar))
	if code != exitClean {
		t.Fatalf("expected exit %d, got %d", exitClean, code)
	}
	if stdout != "" {
		t.Fatalf("expected no output on standard output, got %q", stdout)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read the output file: %v", err)
	}
	if string(output) != testMinifiedGrammar {
		t.Fatalf("expected %q, got %q", testMinifiedGrammar, string(output))
	}
}

func TestLintText(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	stdout, code := runMain(t, testGrammar, "lint")
	if code != exitReported {
		t.Fatalf("expected exit %d, got %d", exitReported, code)
	}

	for _, expected := range []string{
		"stdin:1:1:",
		"removable-comment",
		"redundant-whitespace",
		`minification writes it as "root=(\"a\"/\"b\") DIGIT"`,
		"3 findings",
	} {
		if !strings.Contains(stdout, expected) {
			t.Fatalf("expected the report to hold %q, got:\n%s", expected, stdout)
		}
	}
}

// TestLintNamesTheLeadingCheck checks that a rule whose only fault is one
// kind is reported under the check for that kind, rather than under a
// catch-all.
func TestLintNamesTheLeadingCheck(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "repeat", input: "root=1*1DIGIT\r\n", expected: "(redundant-repeat)"},
		{name: "literal", input: "root=%i\"a\"\r\n", expected: "(non-canonical-literal)"},
		{name: "line ending", input: "root=\"a\"\n", expected: "(non-crlf-line-ending)"},
	}

	for _, testCase := range testCases { //nolint:paralleltest // shares process-global state through run()
		t.Run(testCase.name, func(t *testing.T) {
			stdout, code := runMain(t, testCase.input, "lint")
			if code != exitReported {
				t.Fatalf("expected exit %d, got %d", exitReported, code)
			}
			if !strings.Contains(stdout, testCase.expected) {
				t.Fatalf("expected the report to hold %q, got:\n%s", testCase.expected, stdout)
			}
		})
	}
}

func TestLintClean(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	// A minified definition whose every rule is referred to draws nothing.
	stdout, code := runMain(t, "root=other\r\nother=\"a\"/root\r\n", "lint")
	if code != exitClean {
		t.Fatalf("expected exit %d, got %d: %s", exitClean, code, stdout)
	}
	if stdout != "" {
		t.Fatalf("expected no report, got %q", stdout)
	}
}

func TestLintSarif(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	stdout, code := runMain(t, testGrammar, "lint", "--format", "sarif")
	if code != exitReported {
		t.Fatalf("expected exit %d, got %d", exitReported, code)
	}

	var log struct {
		Schema  string `json:"$schema"`
		Version string `json:"version"`
		Runs    []struct {
			Tool struct {
				Driver struct {
					Name  string `json:"name"`
					Rules []struct {
						Id string `json:"id"`
					} `json:"rules"`
				} `json:"driver"`
			} `json:"tool"`
			Results []struct {
				RuleId    string `json:"ruleId"`
				Locations []struct {
					PhysicalLocation struct {
						ArtifactLocation struct {
							Uri string `json:"uri"`
						} `json:"artifactLocation"`
						Region struct {
							StartLine int `json:"startLine"`
						} `json:"region"`
					} `json:"physicalLocation"`
				} `json:"locations"`
				Fixes []struct {
					ArtifactChanges []struct {
						Replacements []struct {
							InsertedContent struct {
								Text string `json:"text"`
							} `json:"insertedContent"`
						} `json:"replacements"`
					} `json:"artifactChanges"`
				} `json:"fixes"`
			} `json:"results"`
		} `json:"runs"`
	}

	if err := json.Unmarshal([]byte(stdout), &log); err != nil {
		t.Fatalf("json unmarshal: %v\n%s", err, stdout)
	}

	if log.Version != "2.1.0" {
		t.Fatalf("expected SARIF 2.1.0, got %q", log.Version)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("expected one run, got %d", len(log.Runs))
	}

	run := log.Runs[0]
	if run.Tool.Driver.Name != "abnf" {
		t.Fatalf("expected the driver to be abnf, got %q", run.Tool.Driver.Name)
	}
	if len(run.Tool.Driver.Rules) == 0 {
		t.Fatal("expected the driver to declare its rules")
	}
	if len(run.Results) == 0 {
		t.Fatal("expected results")
	}

	fixes := 0
	for _, result := range run.Results {
		if len(result.Locations) == 0 {
			t.Fatalf("the result for %s has no location", result.RuleId)
		}
		if uri := result.Locations[0].PhysicalLocation.ArtifactLocation.Uri; uri != stdinName {
			t.Fatalf("expected the location to be %q, got %q", stdinName, uri)
		}
		if line := result.Locations[0].PhysicalLocation.Region.StartLine; line < 1 {
			t.Fatalf("expected a one-based line, got %d", line)
		}
		fixes += len(result.Fixes)
	}
	if fixes == 0 {
		t.Fatal("expected at least one fix")
	}
}

func TestFix(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	path := writeGrammar(t, testGrammar)

	// Without -write, the command only names what would change.
	stdout, code := runMain(t, "", "fix", path)
	if code != exitReported {
		t.Fatalf("expected exit %d, got %d", exitReported, code)
	}
	if strings.TrimSpace(stdout) != path {
		t.Fatalf("expected the path to be named, got %q", stdout)
	}

	unchanged, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read the grammar file: %v", err)
	}
	if string(unchanged) != testGrammar {
		t.Fatalf("expected the definition to be left alone, got %q", string(unchanged))
	}

	// With -write, it rewrites the definition in place.
	stdout, code = runMain(t, "", "fix", "--write", path)
	if code != exitClean {
		t.Fatalf("expected exit %d, got %d", exitClean, code)
	}
	if stdout != "" {
		t.Fatalf("expected no output, got %q", stdout)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read the grammar file: %v", err)
	}
	if string(written) != testMinifiedGrammar {
		t.Fatalf("expected %q, got %q", testMinifiedGrammar, string(written))
	}

	// A definition already minified is left alone, and reported as such.
	if _, code = runMain(t, "", "fix", path); code != exitClean {
		t.Fatalf("expected exit %d for a minified definition, got %d", exitClean, code)
	}
}

func TestFixKeepsPermissions(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	path := writeGrammar(t, testGrammar)
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatalf("failed to change the permissions: %v", err)
	}

	if _, code := runMain(t, "", "fix", "--write", path); code != exitClean {
		t.Fatalf("expected exit %d, got %d", exitClean, code)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat the grammar file: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o640 {
		t.Fatalf("expected the permissions to be kept as 0640, got %04o", mode)
	}
}

// TestClusteredShortOptions checks the GNU convention that stdlib flag does
// not follow: one dash introduces short options, which may be clustered.
func TestClusteredShortOptions(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	stdout, code := runMain(t, testGrammar, "minify", "-se")
	if code != exitClean {
		t.Fatalf("expected exit %d, got %d", exitClean, code)
	}
	if stdout != testSimplifiedGrammar {
		t.Fatalf("expected %q, got %q", testSimplifiedGrammar, stdout)
	}
}

// TestSingleDashLongNameIsRejected checks the other half of that convention:
// one dash is not two, so a long name behind one dash is not accepted.
func TestSingleDashLongNameIsRejected(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	if _, code := runMain(t, testGrammar, "minify", "-simplify"); code != exitError {
		t.Fatalf("expected exit %d, got %d", exitError, code)
	}
}

// TestAbbreviatedLongName checks that an unambiguous prefix of a long name
// stands for the whole name, as argparse and getopt_long have it.
func TestAbbreviatedLongName(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	stdout, code := runMain(t, testGrammar, "minify", "--simp")
	if code != exitClean {
		t.Fatalf("expected exit %d, got %d", exitClean, code)
	}
	if stdout != testSimplifiedGrammar {
		t.Fatalf("expected %q, got %q", testSimplifiedGrammar, stdout)
	}
}

func TestUsage(t *testing.T) { //nolint:paralleltest // shares process-global state through run()
	testCases := []struct {
		name      string
		arguments []string
		expected  int
	}{
		{name: "no command", expected: exitError},
		{name: "unknown command", arguments: []string{"frobnicate"}, expected: exitError},
		{name: "unknown option", arguments: []string{"lint", "--frobnicate"}, expected: exitError},

		{name: "unknown format", arguments: []string{"lint", "--format", "yaml"}, expected: exitError},
		{name: "fix without paths", arguments: []string{"fix"}, expected: exitError},
		{name: "minify with several paths", arguments: []string{"minify", "a", "b"}, expected: exitError},
		{name: "help", arguments: []string{"--help"}, expected: exitClean},
		{name: "short help", arguments: []string{"-h"}, expected: exitClean},
		{name: "subcommand help", arguments: []string{"lint", "--help"}, expected: exitClean},
	}

	for _, testCase := range testCases { //nolint:paralleltest // shares process-global state through run()
		t.Run(testCase.name, func(t *testing.T) {
			if _, code := runMain(t, "", testCase.arguments...); code != testCase.expected {
				t.Fatalf("expected exit %d, got %d", testCase.expected, code)
			}
		})
	}
}
