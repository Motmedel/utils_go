package content_security_policy

// Report is the deprecated report body sent via the report-uri directive
// (CSP Level 3 section 5.3). JSON keys use hyphenated names.
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

// CSPViolationReportBody is the report body for "csp-violation" reports sent
// via the Reporting API (report-to directive). Defined in CSP Level 3 section 5.
type CSPViolationReportBody struct {
	DocumentURL        string `json:"documentURL,omitempty"`
	Referrer           string `json:"referrer,omitempty"`
	BlockedURL         string `json:"blockedURL,omitempty"`
	EffectiveDirective string `json:"effectiveDirective,omitempty"`
	OriginalPolicy     string `json:"originalPolicy,omitempty"`
	SourceFile         string `json:"sourceFile,omitempty"`
	Sample             string `json:"sample,omitempty"`
	Disposition        string `json:"disposition,omitempty"`
	StatusCode         int    `json:"statusCode,omitempty"`
	LineNumber         *int   `json:"lineNumber,omitempty"`
	ColumnNumber       *int   `json:"columnNumber,omitempty"`
}

// CSPHashReportBody is the report body for "csp-hash" reports sent via the
// Reporting API. Defined in CSP Level 3 section 4.1.4.
type CSPHashReportBody struct {
	DocumentURL    string `json:"documentURL,omitempty"`
	SubresourceURL string `json:"subresourceURL,omitempty"`
	Hash           string `json:"hash,omitempty"`
	Destination    string `json:"destination,omitempty"`
	Type           string `json:"type,omitempty"`
}
