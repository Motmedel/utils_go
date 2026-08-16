package abnf

import (
	"errors"
	"slices"
	"sort"
	"strings"
	"testing"
)

const testGrammar = "root = key \"=\" value\r\nkey = 1*ALPHA\r\nvalue = 1*(ALPHA / DIGIT)\r\n"

// isErrorType reports whether an error in err's tree is of type T.
func isErrorType[T error](err error) bool {
	_, ok := errors.AsType[T](err)
	return ok
}

func TestParseABNF(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		input        string
		errorMatcher func(error) bool
	}{
		{name: "valid grammar", input: testGrammar},
		{
			name:  "comments and empty lines",
			input: "; a comment\r\nroot = \"a\" ; trailing comment\r\n\r\n",
		},
		{name: "incremental alternative", input: "root = \"a\"\r\nroot =/ \"b\"\r\n"},
		{
			name:  "invalid grammar",
			input: "not a grammar",
			errorMatcher: func(err error) bool {
				return errors.Is(err, ErrNoSolutionFound)
			},
		},
		{
			name:  "empty grammar",
			input: "",
			errorMatcher: func(err error) bool {
				return errors.Is(err, ErrNoSolutionFound)
			},
		},
		{
			name:  "missing final crlf",
			input: "root = \"a\"",
			errorMatcher: func(err error) bool {
				return errors.Is(err, ErrNoSolutionFound)
			},
		},
		{
			name:         "duplicated rule",
			input:        "root = \"a\"\r\nroot = \"b\"\r\n",
			errorMatcher: isErrorType[*DuplicatedRuleError],
		},
		{
			name:         "incremental alternative without base rule",
			input:        "root =/ \"a\"\r\n",
			errorMatcher: isErrorType[*RuleNotFoundError],
		},
		{
			name:         "core rule modification",
			input:        "ALPHA = \"a\"\r\n",
			errorMatcher: isErrorType[*CoreRuleModificationError],
		},
		{
			name:         "missing dependency",
			input:        "root = missing-rule\r\n",
			errorMatcher: isErrorType[*DependencyNotFoundError],
		},
		{
			name:         "invalid repetition",
			input:        "root = 4*2\"a\"\r\n",
			errorMatcher: isErrorType[*InvalidRepetitionError],
		},
		{
			name:         "too large num-val",
			input:        "root = %x110000\r\n",
			errorMatcher: isErrorType[*InvalidNumeralError],
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			grammar, err := ParseABNF([]byte(testCase.input))
			if errorMatcher := testCase.errorMatcher; errorMatcher != nil {
				if err == nil {
					t.Fatal("expected an error")
				}
				if !errorMatcher(err) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse abnf: %v", err)
			}
			if grammar == nil {
				t.Fatal("nil grammar")
			}
		})
	}
}

func TestParse(t *testing.T) {
	t.Parallel()

	grammar, err := ParseABNF([]byte(testGrammar))
	if err != nil {
		t.Fatalf("parse abnf: %v", err)
	}

	testCases := []struct {
		name         string
		input        string
		rootRulename string
		numPaths     int
		expectError  bool
	}{
		{name: "valid input", input: "a=1", rootRulename: "root", numPaths: 1},
		{name: "valid alphabetic value", input: "abc=def", rootRulename: "root", numPaths: 1},
		{name: "case-insensitive rulename", input: "a=1", rootRulename: "ROOT", numPaths: 1},
		{name: "invalid input", input: "=1", rootRulename: "root", numPaths: 0},
		{name: "partial input", input: "a=", rootRulename: "root", numPaths: 0},
		{name: "trailing input", input: "a=1 ", rootRulename: "root", numPaths: 0},
		{name: "core rule as root", input: "a", rootRulename: "ALPHA", numPaths: 1},
		{name: "unknown root rule", input: "a=1", rootRulename: "missing", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			paths, err := Parse([]byte(testCase.input), grammar, testCase.rootRulename)
			if testCase.expectError {
				if err == nil {
					t.Fatal("expected an error")
				}
				if _, ok := errors.AsType[*RuleNotFoundError](err); !ok {
					t.Fatalf("expected a rule-not-found error, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if len(paths) != testCase.numPaths {
				t.Fatalf("expected %d paths, got %d", testCase.numPaths, len(paths))
			}
		})
	}
}

func TestParseMatching(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		grammar string
		input   string
		matches bool
	}{
		{name: "char-val insensitive lower", grammar: "root = \"aB\"\r\n", input: "ab", matches: true},
		{name: "char-val insensitive upper", grammar: "root = \"aB\"\r\n", input: "AB", matches: true},
		{name: "char-val sensitive match", grammar: "root = %s\"aB\"\r\n", input: "aB", matches: true},
		{name: "char-val sensitive mismatch", grammar: "root = %s\"aB\"\r\n", input: "ab", matches: false},
		{name: "num-val series match", grammar: "root = \"a\" %x0D.0A\r\n", input: "a\r\n", matches: true},
		{name: "num-val series mismatch", grammar: "root = \"a\" %x0D.0A\r\n", input: "a\r", matches: false},
		{name: "num-val range match", grammar: "root = 1*%d48-57\r\n", input: "0159", matches: true},
		{name: "num-val range mismatch", grammar: "root = 1*%d48-57\r\n", input: "015a", matches: false},
		{name: "binary num-val", grammar: "root = %b1000001\r\n", input: "A", matches: true},
		{
			name:    "multi-byte range match",
			grammar: "root = 1*%x80-10FFFF\r\n",
			input:   "éö🙂",
			matches: true,
		},
		{
			name:    "multi-byte range mismatch",
			grammar: "root = 1*%x80-10FFFF\r\n",
			input:   "é7",
			matches: false,
		},
		{
			name:    "char-val after multi-byte input",
			grammar: "root = %x80-10FFFF \"a\"\r\n",
			input:   "éa",
			matches: true,
		},
		{name: "option present", grammar: "root = \"a\" [\"b\"] \"c\"\r\n", input: "abc", matches: true},
		{name: "option absent", grammar: "root = \"a\" [\"b\"] \"c\"\r\n", input: "ac", matches: true},
		{name: "group alternation", grammar: "root = (\"a\" / \"b\") \"c\"\r\n", input: "bc", matches: true},
		{name: "bounded repetition match", grammar: "root = 2*3\"a\"\r\n", input: "aaa", matches: true},
		{name: "bounded repetition under", grammar: "root = 2*3\"a\"\r\n", input: "a", matches: false},
		{name: "bounded repetition over", grammar: "root = 2*3\"a\"\r\n", input: "aaaa", matches: false},
		{name: "exact repetition", grammar: "root = 2\"a\"\r\n", input: "aa", matches: true},
		{
			name:    "incremental alternative",
			grammar: "root = \"a\"\r\nroot =/ \"b\"\r\n",
			input:   "b",
			matches: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			grammar, err := ParseABNF([]byte(testCase.grammar))
			if err != nil {
				t.Fatalf("parse abnf: %v", err)
			}

			paths, err := Parse([]byte(testCase.input), grammar, "root")
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if matches := len(paths) != 0; matches != testCase.matches {
				t.Fatalf("expected matches=%t, got matches=%t", testCase.matches, matches)
			}
		})
	}
}

// sortedRuleStrings returns the canonical rule representations of a
// grammar, sorted for comparison.
func sortedRuleStrings(grammar *Grammar) []string {
	var ruleStrings []string
	for _, rule := range grammar.Rulemap {
		ruleStrings = append(ruleStrings, rule.String())
	}
	sort.Strings(ruleStrings)
	return ruleStrings
}

func TestGrammarStringRoundTrip(t *testing.T) {
	t.Parallel()

	grammarTexts := []string{
		testGrammar,
		"root = 1*(\"a\" / %x30-39) [\"!\" 2*4%b101 <prose>]\r\nother = %s\"Xy\" %d13.10 root\r\n",
	}

	for _, grammarText := range grammarTexts {
		grammar, err := ParseABNF([]byte(grammarText))
		if err != nil {
			t.Fatalf("parse abnf: %v", err)
		}

		reparsedGrammar, err := ParseABNF([]byte(grammar.String()))
		if err != nil {
			t.Fatalf("parse abnf (round trip): %v", err)
		}

		if !slices.Equal(sortedRuleStrings(grammar), sortedRuleStrings(reparsedGrammar)) {
			t.Fatalf(
				"round trip mismatch:\noriginal: %q\nreparsed: %q",
				sortedRuleStrings(grammar),
				sortedRuleStrings(reparsedGrammar),
			)
		}
	}
}

func TestBootstrapGrammarRoundTrip(t *testing.T) {
	t.Parallel()

	reparsedGrammar, err := ParseABNF([]byte(abnfGrammar.String()))
	if err != nil {
		t.Fatalf("parse abnf (bootstrap): %v", err)
	}

	if !slices.Equal(sortedRuleStrings(abnfGrammar), sortedRuleStrings(reparsedGrammar)) {
		t.Fatal("bootstrap grammar round trip mismatch")
	}
}

// listGrammar is the example RFC 7230 Section 7 gives for the list operator,
// with the OWS its expansion separates elements with.
const listGrammar = "example-list=1#example-list-elmt\r\n" +
	"example-list-elmt=token\r\n" +
	"token=1*(ALPHA/DIGIT)\r\n" +
	"OWS=*(SP/HTAB)\r\n"

// TestParseListExtension checks the "#" list operator of RFC 9110
// Section 5.6.1 against the values RFC 7230 Section 7 calls valid and
// invalid for its own example.
func TestParseListExtension(t *testing.T) {
	t.Parallel()

	grammar, err := ParseABNF([]byte(listGrammar))
	if err != nil {
		t.Fatalf("parse abnf: %v", err)
	}

	testCases := []struct {
		name    string
		input   string
		matches bool
	}{
		{name: "two elements", input: "foo,bar", matches: true},
		// A recipient "MUST parse and ignore" empty list elements.
		{name: "empty element and trailing comma", input: "foo ,bar,", matches: true},
		{name: "several empty elements", input: "foo , ,bar,charlie", matches: true},
		{name: "leading comma", input: ",foo", matches: true},
		// At least one non-empty element is required.
		{name: "empty", input: "", matches: false},
		{name: "comma alone", input: ",", matches: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			paths, err := Parse([]byte(testCase.input), grammar, "example-list")
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if matches := len(paths) != 0; matches != testCase.matches {
				t.Fatalf("expected matches=%t, got matches=%t", testCase.matches, matches)
			}
		})
	}
}

// TestListExtensionRoundTrips checks that the list operator survives being
// read: it is kept as written, and the standard ABNF it stands for is
// carried alongside it.
func TestListExtensionRoundTrips(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		rule string
		// expansion is the standard ABNF of RFC 7230 Section 7.
		expansion string
	}{
		{
			name:      "optional list",
			rule:      "root = #token",
			expansion: `[("," / token) *(OWS "," [OWS token])]`,
		},
		{
			name:      "one or more",
			rule:      "root = 1#token",
			expansion: `*("," OWS) token *(OWS "," [OWS token])`,
		},
		{
			name:      "bounded",
			rule:      "root = 2#5token",
			expansion: `token 1*4(OWS "," OWS token)`,
		},
		{
			name:      "minimum only",
			rule:      "root = 3#token",
			expansion: `token 2*(OWS "," OWS token)`,
		},
		{
			name:      "maximum only",
			rule:      "root = #3token",
			expansion: `[token *2(OWS "," OWS token)]`,
		},
		{
			name:      "group element",
			rule:      "root = 1#(token token)",
			expansion: `*("," OWS) (token token) *(OWS "," [OWS (token token)])`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			grammar, err := ParseABNF([]byte(testCase.rule + "\r\ntoken=ALPHA\r\nOWS=*(SP/HTAB)\r\n"))
			if err != nil {
				t.Fatalf("parse abnf: %v", err)
			}

			rule := grammar.Rule("root")
			if rule == nil {
				t.Fatal("no root rule")
			}

			// The operator is kept as written, so that a definition using it
			// stays as short as it was.
			if written := rule.String(); written != testCase.rule {
				t.Fatalf("expected %q, got %q", testCase.rule, written)
			}

			listElement := soleListElement(t, rule)
			if expansion := listElement.Expansion.String(); expansion != testCase.expansion {
				t.Fatalf("expected the expansion %q, got %q", testCase.expansion, expansion)
			}
		})
	}
}

// soleListElement returns the list element a rule is made of.
func soleListElement(t *testing.T, rule *Rule) *ListElement {
	t.Helper()

	repetitions := rule.Alternation.Concatenations[0].Repetitions
	if len(repetitions) != 1 {
		t.Fatalf("expected one repetition, got %d", len(repetitions))
	}

	listElement, ok := repetitions[0].Element.(*ListElement)
	if !ok {
		t.Fatalf("expected a list element, got %T", repetitions[0].Element)
	}

	return listElement
}

// TestListExtensionNeedsSeparator checks that the rule the expansion refers
// to has to be defined, as any other referenced rule does.
func TestListExtensionNeedsSeparator(t *testing.T) {
	t.Parallel()

	_, err := ParseABNF([]byte("root=1#token\r\ntoken=ALPHA\r\n"))
	if err == nil {
		t.Fatal("expected an error")
	}

	notFound, ok := errors.AsType[*DependencyNotFoundError](err)
	if !ok {
		t.Fatalf("expected a dependency-not-found error, got: %v", err)
	}
	// Dependency names are reported lowercased, as rule names are
	// case-insensitive.
	if !strings.EqualFold(notFound.Rulename, "OWS") {
		t.Fatalf("expected the missing dependency to be OWS, got %q", notFound.Rulename)
	}
}

// TestListExtensionCounts checks that the counts on either side of the "#"
// are read as RFC 7230 Section 7 gives them, with an absent minimum meaning
// none and an absent maximum meaning unbounded.
func TestListExtensionCounts(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		rule    string
		minimum int
		maximum int
	}{
		{name: "neither", rule: "root = #token", minimum: 0, maximum: Inf},
		{name: "minimum", rule: "root = 1#token", minimum: 1, maximum: Inf},
		{name: "maximum", rule: "root = #5token", minimum: 0, maximum: 5},
		{name: "both", rule: "root = 2#5token", minimum: 2, maximum: 5},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			grammar, err := ParseABNF([]byte(testCase.rule + "\r\ntoken=ALPHA\r\nOWS=*(SP/HTAB)\r\n"))
			if err != nil {
				t.Fatalf("parse abnf: %v", err)
			}

			listElement := soleListElement(t, grammar.Rule("root"))
			if listElement.Min != testCase.minimum || listElement.Max != testCase.maximum {
				t.Fatalf(
					"expected %d to %d, got %d to %d",
					testCase.minimum,
					testCase.maximum,
					listElement.Min,
					listElement.Max,
				)
			}
		})
	}
}
