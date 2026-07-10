package role_privilege

type RolePrivilege struct {
	PrivilegeName string `json:"privilegeName,omitempty"`
	ServiceId     string `json:"serviceId,omitempty"`
}
