package abnf

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Inf marks an unbounded repetition maximum.
const Inf = -1

// Element is the interface implemented by all ABNF element variants:
// *RulenameElement, *GroupElement, *OptionElement, *CharValElement,
// *NumValElement and *ProseValElement.
type Element interface {
	fmt.Stringer

	element()
}

// Rule is an ABNF rule: a unique name and its defining alternation.
type Rule struct {
	// Name is the rule name, case-insensitive according to RFC 5234
	// Section 2.1.
	Name        string
	Alternation *Alternation
}

func (rule *Rule) String() string {
	return fmt.Sprintf("%s = %s", rule.Name, rule.Alternation)
}

// Alternation is an ABNF alternation: a choice between concatenations.
type Alternation struct {
	Concatenations []*Concatenation
}

func (alternation *Alternation) String() string {
	parts := make([]string, len(alternation.Concatenations))
	for i, concatenation := range alternation.Concatenations {
		parts[i] = concatenation.String()
	}
	return strings.Join(parts, " / ")
}

// Concatenation is an ABNF concatenation: an ordered sequence of
// repetitions.
type Concatenation struct {
	Repetitions []*Repetition
}

func (concatenation *Concatenation) String() string {
	parts := make([]string, len(concatenation.Repetitions))
	for i, repetition := range concatenation.Repetitions {
		parts[i] = repetition.String()
	}
	return strings.Join(parts, " ")
}

// Repetition is an ABNF repetition of an element. Max may be Inf (-1) for
// an unbounded repetition.
type Repetition struct {
	Min, Max int
	Element  Element
}

func (repetition *Repetition) String() string {
	if repetition.Min == repetition.Max {
		if repetition.Min == 1 {
			return repetition.Element.String()
		}
		return strconv.Itoa(repetition.Min) + repetition.Element.String()
	}
	str := ""
	if repetition.Min != 0 {
		str += strconv.Itoa(repetition.Min)
	}
	str += "*"
	if repetition.Max != Inf {
		str += strconv.Itoa(repetition.Max)
	}
	return str + repetition.Element.String()
}

// RulenameElement is an ABNF rulename element: a reference to another rule.
type RulenameElement struct {
	// Name is the referenced rule name, case-insensitive according to
	// RFC 5234 Section 2.1.
	Name string
}

func (element *RulenameElement) String() string {
	return element.Name
}

func (element *RulenameElement) element() {}

var _ Element = (*RulenameElement)(nil)

// GroupElement is an ABNF group element: a parenthesised alternation.
type GroupElement struct {
	Alternation *Alternation
}

func (element *GroupElement) String() string {
	return "(" + element.Alternation.String() + ")"
}

func (element *GroupElement) element() {}

var _ Element = (*GroupElement)(nil)

// OptionElement is an ABNF option element: a bracketed alternation,
// equivalent to a 0*1 repetition of it.
type OptionElement struct {
	Alternation *Alternation
}

func (element *OptionElement) String() string {
	return "[" + element.Alternation.String() + "]"
}

func (element *OptionElement) element() {}

var _ Element = (*OptionElement)(nil)

// CharValElement is an ABNF char-val element: a literal string of printable
// US-ASCII characters. Matching is case-insensitive unless Sensitive is set
// (the RFC 7405 "%s" form).
type CharValElement struct {
	Sensitive bool
	Value     string
}

func (element *CharValElement) String() string {
	if element.Sensitive {
		return `%s"` + element.Value + `"`
	}
	return `"` + element.Value + `"`
}

func (element *CharValElement) element() {}

var _ Element = (*CharValElement)(nil)

// NumValStatus defines how the values of a NumValElement combine.
type NumValStatus int

const (
	// NumValStatusSeries matches each value in order.
	NumValStatusSeries NumValStatus = iota
	// NumValStatusRange matches any value within the two bounds.
	NumValStatusRange
)

// NumValElement is an ABNF num-val element: numeric character values in a
// given base, either as a series or as a range.
type NumValElement struct {
	// Base is the num-val base character: "b", "d" or "x".
	Base   string
	Status NumValStatus
	Values []string
}

func (element *NumValElement) String() string {
	separator := "."
	if element.Status == NumValStatusRange {
		separator = "-"
	}
	return "%" + element.Base + strings.Join(element.Values, separator)
}

func (element *NumValElement) element() {}

var _ Element = (*NumValElement)(nil)

// ListElement is the "#" list operator of RFC 9110 Section 5.6.1: a list of
// Min to Max occurrences of an element, separated by commas and optional
// whitespace. Max may be Inf.
//
// RFC 5234 defines no such operator. The HTTP specifications define it in
// terms of standard ABNF, and Expansion holds that standard ABNF, which is
// what matching against the element uses.
type ListElement struct {
	Min, Max int
	Element  Element
	// Expansion is the standard ABNF the operator stands for, as
	// RFC 7230 Section 7 gives it. It holds Element itself rather than a
	// copy of it, so rewriting an element of a parsed grammar in place
	// rewrites it here too; treat a list element as a whole.
	Expansion *Alternation
}

func (element *ListElement) String() string {
	str := ""
	if element.Min != 0 {
		str += strconv.Itoa(element.Min)
	}
	str += "#"
	if element.Max != Inf {
		str += strconv.Itoa(element.Max)
	}
	return str + element.Element.String()
}

func (element *ListElement) element() {}

// alternation returns the standard ABNF the operator stands for, building it
// if the element was made without it.
func (element *ListElement) alternation() *Alternation {
	if element.Expansion != nil {
		return element.Expansion
	}
	return expandList(element.Min, element.Max, element.Element)
}

var _ Element = (*ListElement)(nil)

// ProseValElement is an ABNF prose-val element: a free-form prose
// description that cannot be matched against input.
type ProseValElement struct {
	Value string
}

func (element *ProseValElement) String() string {
	return "<" + element.Value + ">"
}

func (element *ProseValElement) element() {}

var _ Element = (*ProseValElement)(nil)

// ParseNumVal converts a num-val numeral in the given base into the
// corresponding rune. Values beyond the maximum Unicode code point are
// rejected.
func ParseNumVal(value, base string) (rune, error) {
	var intBase int
	switch base {
	case "B", "b":
		intBase = 2
	case "D", "d":
		intBase = 10
	case "X", "x":
		intBase = 16
	default:
		return 0, &InvalidNumeralError{Base: base, Value: value}
	}

	number, err := strconv.ParseUint(value, intBase, 32)
	if err != nil || number > unicode.MaxRune {
		return 0, &InvalidNumeralError{Base: base, Value: value}
	}

	return rune(number), nil
}

// mustParseNumVal converts a num-val numeral in the given base into the
// corresponding rune, panicking on invalid input. It is reserved for
// num-vals of semantically validated grammars, for which conversion cannot
// fail.
func mustParseNumVal(value, base string) rune {
	r, err := ParseNumVal(value, base)
	if err != nil {
		panic(err)
	}
	return r
}
