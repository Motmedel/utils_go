package abnf

import (
	"errors"
	"fmt"
)

// ErrNoSolutionFound is returned when parsing an ABNF grammar definition
// yields no solution, meaning the definition is invalid.
var ErrNoSolutionFound = errors.New("no solution found, the input ABNF grammar may be invalid")

// errUnexpectedPathStructure is returned when a parsed path does not have
// the structure the evaluator expects; this indicates a bug rather than
// invalid input.
var errUnexpectedPathStructure = errors.New("unexpected path structure")

// MultipleSolutionsFoundError is returned when parsing found multiple
// solutions where at most one was expected.
type MultipleSolutionsFoundError struct {
	Paths []*Path
}

func (multipleSolutionsFoundError *MultipleSolutionsFoundError) Error() string {
	return "multiple solutions found when at most one was expected"
}

var _ error = (*MultipleSolutionsFoundError)(nil)

// RuleNotFoundError is returned when a rule is not part of the grammar.
type RuleNotFoundError struct {
	Rulename string
}

func (ruleNotFoundError *RuleNotFoundError) Error() string {
	return fmt.Sprintf("rule %q was not found in the grammar", ruleNotFoundError.Rulename)
}

var _ error = (*RuleNotFoundError)(nil)

// CoreRuleModificationError is returned when a grammar attempts to redefine
// or extend a core rule.
type CoreRuleModificationError struct {
	Rulename string
}

func (coreRuleModificationError *CoreRuleModificationError) Error() string {
	return fmt.Sprintf("core rule %q cannot be modified", coreRuleModificationError.Rulename)
}

var _ error = (*CoreRuleModificationError)(nil)

// DuplicatedRuleError is returned when a grammar defines the same rule more
// than once.
type DuplicatedRuleError struct {
	Rulename string
}

func (duplicatedRuleError *DuplicatedRuleError) Error() string {
	return fmt.Sprintf("rule %q is already defined in the grammar", duplicatedRuleError.Rulename)
}

var _ error = (*DuplicatedRuleError)(nil)

// DependencyNotFoundError is returned during semantic validation when a rule
// references a rule that does not exist.
type DependencyNotFoundError struct {
	Rulename string
}

func (dependencyNotFoundError *DependencyNotFoundError) Error() string {
	return fmt.Sprintf("unsatisfied rule dependency %q", dependencyNotFoundError.Rulename)
}

var _ error = (*DependencyNotFoundError)(nil)

// InvalidRepetitionError is returned during semantic validation when a
// repetition has a minimum greater than its maximum.
type InvalidRepetitionError struct {
	Repetition *Repetition
}

func (invalidRepetitionError *InvalidRepetitionError) Error() string {
	return fmt.Sprintf("invalid repetition %q: the minimum is greater than the maximum", invalidRepetitionError.Repetition)
}

var _ error = (*InvalidRepetitionError)(nil)

// InvalidNumeralError is returned when a num-val numeral cannot be
// represented as a Unicode code point in its base.
type InvalidNumeralError struct {
	Base, Value string
}

func (invalidNumeralError *InvalidNumeralError) Error() string {
	return fmt.Sprintf("invalid numeral value %q for base %q", invalidNumeralError.Value, invalidNumeralError.Base)
}

var _ error = (*InvalidNumeralError)(nil)
