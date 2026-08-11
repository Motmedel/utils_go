package permission

const (
	TypeUser   = "user"
	TypeGroup  = "group"
	TypeDomain = "domain"
	TypeAnyone = "anyone"
)

const (
	RoleOwner         = "owner"
	RoleOrganizer     = "organizer"
	RoleFileOrganizer = "fileOrganizer"
	RoleWriter        = "writer"
	RoleCommenter     = "commenter"
	RoleReader        = "reader"
)

type Permission struct {
	Id                 string `json:"id,omitzero"`
	Type               string `json:"type,omitzero"`
	EmailAddress       string `json:"emailAddress,omitzero"`
	Domain             string `json:"domain,omitzero"`
	Role               string `json:"role,omitzero"`
	DisplayName        string `json:"displayName,omitzero"`
	AllowFileDiscovery bool   `json:"allowFileDiscovery,omitzero"`
	ExpirationTime     string `json:"expirationTime,omitzero"`
	Deleted            bool   `json:"deleted,omitzero"`
	PendingOwner       bool   `json:"pendingOwner,omitzero"`
	View               string `json:"view,omitzero"`
}
