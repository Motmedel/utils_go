package abnf

import (
	"strings"
)

// Grammar is an ABNF grammar as defined by RFC 5234: a set of uniquely
// named rules.
type Grammar struct {
	// Rules holds the rules in definition order.
	Rules []*Rule
	// Rulemap indexes the rules of Rules by lowercase rule name.
	Rulemap map[string]*Rule
}

// String returns an ABNF representation of the grammar, using CRLF line
// endings as the specification requires. The rules appear in definition
// order.
func (grammar *Grammar) String() string {
	var sb strings.Builder
	for _, rule := range grammar.Rules {
		sb.WriteString(rule.String())
		sb.WriteString("\r\n")
	}
	return sb.String()
}

// Rule returns the rule with the given name, falling back to the core rules
// of RFC 5234 Section 8.1, or nil if neither holds it. Rule names are
// case-insensitive according to RFC 5234 Section 2.1, so the name needs not
// match the spelling the rule was defined with.
func (grammar *Grammar) Rule(rulename string) *Rule {
	return getRule(rulename, grammar.Rulemap)
}

// newGrammar makes a grammar out of rules given in definition order,
// indexing them by lowercase rule name.
func newGrammar(rules ...*Rule) *Grammar {
	rulemap := make(map[string]*Rule, len(rules))
	for _, rule := range rules {
		rulemap[strings.ToLower(rule.Name)] = rule
	}
	return &Grammar{Rules: rules, Rulemap: rulemap}
}

// getRule returns the rule with the given name from the rulemap, falling
// back to the core rules, or nil if not found. Rule names are
// case-insensitive according to RFC 5234 Section 2.1.
func getRule(rulename string, rulemap map[string]*Rule) *Rule {
	key := strings.ToLower(rulename)
	if rule := rulemap[key]; rule != nil {
		return rule
	}
	return coreRules[key]
}
