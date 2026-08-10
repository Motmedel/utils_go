package spf

const Prefix = "v=spf1 "

const MaximumLookupLimit = 10

const (
	NeutralQualifier  = "?"
	SoftfailQualifier = "~"
	FailQualifier     = "-"
)

type Mechanism struct {
	Label string `json:"label,omitzero"`
	Value string `json:"value,omitzero"`
}

type Directive struct {
	Index     int        `json:"index"`
	Qualifier string     `json:"qualifier,omitzero"`
	Mechanism *Mechanism `json:"mechanism,omitzero"`
}

type Modifier struct {
	Index int    `json:"index"`
	Label string `json:"label,omitzero"`
	Value string `json:"value,omitzero"`
}

type TermPtr interface {
	*Modifier | *Directive
}

func getTypedTerms[T TermPtr](record *Record) []T {
	var typedTerms []T

	for _, term := range record.Terms {
		switch typedTerm := term.(type) {
		case T:
			typedTerms = append(typedTerms, typedTerm)
		}
	}

	return typedTerms
}

type Record struct {
	Domain string `json:"domain,omitzero"`
	Raw    string `json:"raw,omitzero"`
	Terms  []any  `json:"-" jsonschema:"-"`
}

func (r *Record) Modifiers() []*Modifier {
	return getTypedTerms[*Modifier](r)
}

func (r *Record) Directives() []*Directive {
	return getTypedTerms[*Directive](r)
}
