package content_security_policy

type Report struct {
	DocumentURI        string `json:"document-uri,omitempty" required:"true"`
	Referrer           string `json:"referrer,omitempty" required:"true"`
	ViolatedDirective  string `json:"violated-directive,omitempty" required:"true"`
	EffectiveDirective string `json:"effective-directive,omitempty"`
	OriginalPolicy     string `json:"original-policy,omitempty" required:"true"`
	BlockedUri         string `json:"blocked-uri,omitempty" required:"true"`
	LineNumber         int    `json:"line-number,omitempty"`
	ColumnNumber       int    `json:"column-number,omitempty"`
	SourceFile         string `json:"source-file,omitempty"`
}
