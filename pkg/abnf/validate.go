package abnf

import (
	"slices"
	"strings"
)

// validateGrammar performs semantic validation of a grammar:
//   - every rule dependency exists,
//   - every repetition has min <= max,
//   - every num-val numeral is a valid Unicode code point in its base.
func validateGrammar(grammar *Grammar) error {
	for _, rule := range grammar.Rulemap {
		for _, dependency := range alternationDependencies(rule.Alternation) {
			if getRule(dependency, grammar.Rulemap) == nil {
				return &DependencyNotFoundError{Rulename: dependency}
			}
		}
	}

	for _, rule := range grammar.Rulemap {
		if err := validateAlternation(rule.Alternation); err != nil {
			return err
		}
	}

	return nil
}

func validateAlternation(alternation *Alternation) error {
	for _, concatenation := range alternation.Concatenations {
		for _, repetition := range concatenation.Repetitions {
			if repetition.Max != inf && repetition.Min > repetition.Max {
				return &InvalidRepetitionError{Repetition: repetition}
			}

			switch element := repetition.Element.(type) {
			case *NumValElement:
				for _, value := range element.Values {
					if _, err := parseNumVal(value, element.Base); err != nil {
						return err
					}
				}
			case *GroupElement:
				if err := validateAlternation(element.Alternation); err != nil {
					return err
				}
			case *OptionElement:
				if err := validateAlternation(element.Alternation); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// alternationDependencies returns the names of the rules the alternation
// references, lowercased and deduplicated.
func alternationDependencies(alternation *Alternation) []string {
	var dependencies []string

	var appendDependencies func(alternation *Alternation)
	appendDependencies = func(alternation *Alternation) {
		for _, concatenation := range alternation.Concatenations {
			for _, repetition := range concatenation.Repetitions {
				switch element := repetition.Element.(type) {
				case *RulenameElement:
					if name := strings.ToLower(element.Name); !slices.Contains(dependencies, name) {
						dependencies = append(dependencies, name)
					}
				case *GroupElement:
					appendDependencies(element.Alternation)
				case *OptionElement:
					appendDependencies(element.Alternation)
				}
			}
		}
	}
	appendDependencies(alternation)

	return dependencies
}
