package abnf

// The builders below construct grammar rules concisely; they are used to
// define the core rules and the ABNF grammar of ABNF in rules.go.

func newRule(name string, concatenations ...*Concatenation) *Rule {
	return &Rule{Name: name, Alternation: alt(concatenations...)}
}

func alt(concatenations ...*Concatenation) *Alternation {
	return &Alternation{Concatenations: concatenations}
}

func cat(repetitions ...*Repetition) *Concatenation {
	return &Concatenation{Repetitions: repetitions}
}

func rep(minCount, maxCount int, element Element) *Repetition {
	return &Repetition{Min: minCount, Max: maxCount, Element: element}
}

func one(element Element) *Repetition {
	return rep(1, 1, element)
}

func ref(name string) *RulenameElement {
	return &RulenameElement{Name: name}
}

func str(value string) *CharValElement {
	return &CharValElement{Value: value}
}

func numRange(base, low, high string) *NumValElement {
	return &NumValElement{Base: base, Status: NumValStatusRange, Values: []string{low, high}}
}

func numSeries(base string, values ...string) *NumValElement {
	return &NumValElement{Base: base, Status: NumValStatusSeries, Values: values}
}

func grp(concatenations ...*Concatenation) *GroupElement {
	return &GroupElement{Alternation: alt(concatenations...)}
}

func opt(concatenations ...*Concatenation) *OptionElement {
	return &OptionElement{Alternation: alt(concatenations...)}
}
