package filter

type SizeComparison string

const (
	SizeComparisonUnspecified SizeComparison = "unspecified"
	SizeComparisonSmaller     SizeComparison = "smaller"
	SizeComparisonLarger      SizeComparison = "larger"
)

type Criteria struct {
	From           string         `json:"from,omitzero"`
	To             string         `json:"to,omitzero"`
	Subject        string         `json:"subject,omitzero"`
	Query          string         `json:"query,omitzero"`
	NegatedQuery   string         `json:"negatedQuery,omitzero"`
	HasAttachment  bool           `json:"hasAttachment,omitzero"`
	ExcludeChats   bool           `json:"excludeChats,omitzero"`
	Size           int            `json:"size,omitzero"`
	SizeComparison SizeComparison `json:"sizeComparison,omitzero"`
}

type Action struct {
	AddLabelIds    []string `json:"addLabelIds,omitzero"`
	RemoveLabelIds []string `json:"removeLabelIds,omitzero"`
	Forward        string   `json:"forward,omitzero"`
}

type Filter struct {
	Id       string    `json:"id,omitzero"`
	Criteria *Criteria `json:"criteria,omitzero"`
	Action   *Action   `json:"action,omitzero"`
}
