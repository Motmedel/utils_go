package credentials_file

type CredentialsFile struct {
	Type           string `json:"type"`
	QuotaProjectId string `json:"quota_project_id"`

	// authorized_user fields
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`

	// service_account fields
	ClientEmail  string `json:"client_email"`
	PrivateKeyID string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	TokenURI     string `json:"token_uri"`
	ProjectID    string `json:"project_id"`
}
