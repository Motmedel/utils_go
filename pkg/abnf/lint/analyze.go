package lint

import (
	"fmt"
	"strings"

	"github.com/Motmedel/utils_go/pkg/abnf"
	"github.com/Motmedel/utils_go/pkg/abnf/minify"
)

// transformRuleIds names the check that each simplifying transform answers
// to.
var transformRuleIds = map[minify.Transform]RuleId{
	minify.TransformGroup:   RuleIdRedundantGroup,
	minify.TransformNumVal:  RuleIdVerboseNumVal,
	minify.TransformLiteral: RuleIdJoinableLiterals,
}

// simplificationFindings reports, for every rule, which transforms would
// shorten it and what they would leave behind. A transform is applied on its
// own so that each finding says what it alone achieves.
func (linter *linter) simplificationFindings() {
	for _, rule := range linter.grammar.Rules {
		paths := linter.rulePaths[strings.ToLower(rule.Name)]
		if len(paths) == 0 {
			continue
		}

		minified := minify.RenderRule(rule)

		for _, transform := range minify.AllTransforms() {
			simplified := minify.RenderRule(rule, transform)
			if simplified == minified {
				continue
			}

			// A rule assembled from incremental alternatives spans more
			// than one definition, so no single replacement expresses it.
			var replacementText *string
			if len(paths) == 1 {
				replacementText = replacement(simplified)
			}

			linter.add(
				linter.source.finding(
					transformRuleIds[transform],
					paths[0].Start,
					paths[0].End,
					fmt.Sprintf(
						"the rule %q can be written as %q",
						strings.TrimSuffix(minified, crlf),
						strings.TrimSuffix(simplified, crlf),
					),
					replacementText,
				),
			)
		}
	}
}

// qualityFindings reports on the grammar itself rather than on how it is
// written.
func (linter *linter) qualityFindings(path *abnf.Path) {
	linter.proseValFindings(path)
	linter.rulenameFindings(path)
	linter.duplicateAlternativeFindings()
	linter.listExtensionFindings(path)
}

// listExtensionFindings reports the rules using the "#" list operator. It is
// read here, and kept as written, but RFC 5234 does not define it: a
// definition holding one goes only where a parser knows the HTTP
// specifications.
func (linter *linter) listExtensionFindings(path *abnf.Path) {
	for _, rulePath := range items(path, nodeRule) {
		if !holdsListRepeat(rulePath, linter.source.normalized) {
			continue
		}

		rulename := item(rulePath, nodeRulename)
		if rulename == nil {
			continue
		}

		linter.add(
			linter.source.finding(
				RuleIdListExtension,
				rulename.Start,
				rulename.End,
				fmt.Sprintf(
					"the rule %q uses the \"#\" list operator, which RFC 9110 Section 5.6.1 defines for "+
						"HTTP field values and RFC 5234 does not define at all",
					linter.source.text(rulename.Start, rulename.End),
				),
				nil,
			),
		)
	}
}

// proseValFindings reports every prose-val. A prose-val describes what to
// match in words instead of matching it, so pkg/abnf finds no path through
// one and any rule needing it can never match.
func (linter *linter) proseValFindings(path *abnf.Path) {
	var walk func(current *abnf.Path)
	walk = func(current *abnf.Path) {
		if current.MatchRule == nodeProseVal {
			linter.add(
				linter.source.finding(
					RuleIdUnmatchableProseVal,
					current.Start,
					current.End,
					fmt.Sprintf(
						"the prose-val %s describes what to match instead of matching it, so nothing ever matches it",
						linter.source.text(current.Start, current.End),
					),
					nil,
				),
			)
			return
		}
		for _, subpath := range current.Subpaths {
			walk(subpath)
		}
	}
	walk(path)
}

// rulenameFindings reports references that spell a rule name with a
// different case than its definition, and rules that nothing refers to.
func (linter *linter) rulenameFindings(path *abnf.Path) {
	referenced := map[string]bool{}

	for _, rulePath := range items(path, nodeRule) {
		rulenames := items(rulePath, nodeRulename)
		if len(rulenames) == 0 {
			continue
		}

		// A rule names itself first, and everything after that is a
		// reference to another rule.
		for _, rulename := range rulenames[1:] {
			name := linter.source.text(rulename.Start, rulename.End)
			referenced[strings.ToLower(name)] = true

			defined := linter.grammar.Rule(name)
			if defined == nil || defined.Name == name {
				continue
			}

			linter.add(
				linter.source.finding(
					RuleIdInconsistentRulenameCase,
					rulename.Start,
					rulename.End,
					fmt.Sprintf("the rule %q is spelled %q where it is defined", name, defined.Name),
					replacement(defined.Name),
				),
			)
		}
	}

	// The list expansion separates elements with OWS without the definition
	// ever spelling it out, so a definition using the operator refers to it
	// all the same.
	if holdsListOperator(linter.source.normalized) {
		referenced[strings.ToLower(listSeparatorRulename)] = true
	}

	for _, rule := range linter.grammar.Rules {
		lowerName := strings.ToLower(rule.Name)
		if referenced[lowerName] {
			continue
		}

		paths := linter.rulePaths[lowerName]
		if len(paths) == 0 {
			continue
		}

		rulename := item(paths[0], nodeRulename)
		if rulename == nil {
			continue
		}

		linter.add(
			linter.source.finding(
				RuleIdUnreferencedRule,
				rulename.Start,
				rulename.End,
				fmt.Sprintf(
					"no rule refers to %q, which is expected only of the rule a grammar is parsed from",
					rule.Name,
				),
				nil,
			),
		)
	}
}

// duplicateAlternativeFindings reports alternations offering the same
// alternative more than once. Every alternative after the first is dead: it
// matches only what the first already matched.
func (linter *linter) duplicateAlternativeFindings() {
	for _, rule := range linter.grammar.Rules {
		paths := linter.rulePaths[strings.ToLower(rule.Name)]
		if len(paths) == 0 {
			continue
		}

		for _, duplicate := range duplicateAlternatives(rule.Alternation) {
			linter.add(
				linter.source.finding(
					RuleIdDuplicateAlternative,
					paths[0].Start,
					paths[0].End,
					fmt.Sprintf(
						"the rule %q offers the alternative %q more than once, and only the first can match",
						rule.Name,
						duplicate,
					),
					nil,
				),
			)
		}
	}
}

// duplicateAlternatives returns the alternatives that an alternation, or any
// alternation within it, offers more than once.
func duplicateAlternatives(alternation *abnf.Alternation) []string {
	if alternation == nil {
		return nil
	}

	var duplicates []string

	seen := map[string]bool{}
	for _, concatenation := range alternation.Concatenations {
		text := concatenation.String()
		if seen[text] {
			duplicates = append(duplicates, text)
		}
		seen[text] = true

		for _, repetition := range concatenation.Repetitions {
			switch element := repetition.Element.(type) {
			case *abnf.GroupElement:
				duplicates = append(duplicates, duplicateAlternatives(element.Alternation)...)
			case *abnf.OptionElement:
				duplicates = append(duplicates, duplicateAlternatives(element.Alternation)...)
			}
		}
	}

	return duplicates
}
