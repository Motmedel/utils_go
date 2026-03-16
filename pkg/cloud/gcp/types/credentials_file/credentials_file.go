package credentials_file

type ServiceAccountImpersonationInfo struct {
	TokenLifetimeSeconds int `json:"token_lifetime_seconds"`
}

type File struct {
	Type string `json:"type"`

	// Service Account fields
	ClientEmail    string `json:"client_email"`
	PrivateKeyID   string `json:"private_key_id"`
	PrivateKey     string `json:"private_key"`
	AuthURL        string `json:"auth_uri"`
	TokenURL       string `json:"token_uri"`
	ProjectID      string `json:"project_id"`
	UniverseDomain string `json:"universe_domain"`

	// User Credential fields
	// (These typically come from gcloud auth.)
	ClientSecret string `json:"client_secret"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`

	// External Account fields
	Audience                       string                           `json:"audience"`
	SubjectTokenType               string                           `json:"subject_token_type"`
	TokenURLExternal               string                           `json:"token_url"`
	TokenInfoURL                   string                           `json:"token_info_url"`
	ServiceAccountImpersonationURL string                           `json:"service_account_impersonation_url"`
	ServiceAccountImpersonation    *ServiceAccountImpersonationInfo `json:"service_account_impersonation"`
	Delegates                      []string                         `json:"delegates"`
	//CredentialSource               externalaccount.CredentialSource `json:"credential_source"`
	QuotaProjectID           string `json:"quota_project_id"`
	WorkforcePoolUserProject string `json:"workforce_pool_user_project"`

	// External Account Authorized User fields
	RevokeURL string `json:"revoke_url"`

	// Service account impersonation
	SourceCredentials *File `json:"source_credentials"`
}
