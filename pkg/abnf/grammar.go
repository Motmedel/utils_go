package abnf

import (
	"strings"
)

// Grammar is an ABNF grammar as defined by RFC 5234: a set of uniquely
// named rules.
type Grammar struct {
	// Rulemap indexes the rules by lowercase rule name.
	Rulemap map[string]*Rule
}

// String returns an ABNF representation of the grammar, using CRLF line
// endings as the specification requires. The rule order is unspecified.
func (grammar *Grammar) String() string {
	var sb strings.Builder
	for _, rule := range grammar.Rulemap {
		sb.WriteString(rule.String())
		sb.WriteString("\r\n")
	}
	return sb.String()
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
