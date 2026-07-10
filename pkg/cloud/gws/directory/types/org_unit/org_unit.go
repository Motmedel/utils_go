package org_unit

type OrgUnit struct {
	Kind string `json:"kind,omitempty"`
	Etag string `json:"etag,omitempty"`

	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	OrgUnitId         string `json:"orgUnitId,omitempty"`
	OrgUnitPath       string `json:"orgUnitPath,omitempty"`
	ParentOrgUnitId   string `json:"parentOrgUnitId,omitempty"`
	ParentOrgUnitPath string `json:"parentOrgUnitPath,omitempty"`

	BlockInheritance bool `json:"blockInheritance,omitempty"`
}
