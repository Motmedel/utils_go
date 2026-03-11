package group

type Group struct {
	Kind string `json:"kind,omitempty"`
	Id   string `json:"id,omitempty"`
	Etag string `json:"etag,omitempty"`

	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	DirectMembersCount string `json:"directMembersCount,omitempty"`
	AdminCreated       bool   `json:"adminCreated,omitempty"`

	Aliases            []string `json:"aliases,omitempty"`
	NonEditableAliases []string `json:"nonEditableAliases,omitempty"`
}
