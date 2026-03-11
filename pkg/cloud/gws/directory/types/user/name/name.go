package name

type Name struct {
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	FullName   string `json:"fullName,omitempty"`
}
