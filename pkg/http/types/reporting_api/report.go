package reporting_api

// Report represents a serialized report as defined in the W3C Reporting API
// specification (section 2.4). The Body field is generic because its structure
// is determined by the report's Type.
type Report[T any] struct {
	Age       int    `json:"age"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	UserAgent string `json:"user_agent"`
	Body      T      `json:"body"`
}
