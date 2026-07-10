package privilege

type Privilege struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`

	ServiceId     string `json:"serviceId,omitempty"`
	ServiceName   string `json:"serviceName,omitempty"`
	PrivilegeName string `json:"privilegeName,omitempty"`
	IsOuScopable  bool   `json:"isOuScopable,omitempty"`

	ChildPrivileges []*Privilege `json:"childPrivileges,omitempty"`
}
