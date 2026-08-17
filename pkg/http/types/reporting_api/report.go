package reporting_api

// Report represents a serialized report as defined in the W3C Reporting API
// specification (section 2.4). The Body field is generic because its structure
// is determined by the report's Type.
type Report[T any] struct {
	// The five members below are what section 2.4 serializes, but a browser is free to send more,
	// and engines do: WebKit adds the report struct's internal "attempts" and "destination", which
	// section 2.1.3 keeps to itself. A schema refusing them would reject the whole batch and lose
	// the reports it was sent, so the object stays open to what an engine adds.
	//nolint:revive // The blank field is what carries what holds for the object; it is read from the type.
	_ struct{} `jsonschema:",additionalProperties:true"`

	Age       int    `json:"age"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	UserAgent string `json:"user_agent" jsonschema:"user_agent,minlength:0"`
	Body      T      `json:"body"`
}
