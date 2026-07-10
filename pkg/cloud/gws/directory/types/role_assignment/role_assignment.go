package role_assignment

type RoleAssignment struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`

	RoleAssignmentId string `json:"roleAssignmentId,omitempty"`
	RoleId           string `json:"roleId,omitempty"`

	AssignedTo   string `json:"assignedTo,omitempty"`
	AssigneeType string `json:"assigneeType,omitempty"`

	ScopeType string `json:"scopeType,omitempty"`
	OrgUnitId string `json:"orgUnitId,omitempty"`
	Condition string `json:"condition,omitempty"`
}
