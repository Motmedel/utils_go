package integrity_policy

import "fmt"

type IntegrityViolationReportBody struct {
	// The body is written by the browser, which is free to say more about what it blocked than
	// what is specified today. What it adds is not worth refusing the report over.
	//nolint:revive // The blank field is what carries what holds for the object; it is read from the type.
	_ struct{} `jsonschema:",additionalProperties:true"`

	DocumentUrl string `json:"documentURL"`
	BlockedUrl  string `json:"blockedURL"`
	Destination string `json:"destination"`
	ReportOnly  bool   `json:"reportOnly"`
}

func (body *IntegrityViolationReportBody) Message() string {
	return fmt.Sprintf(
		"The page's settings blocked a %s at %s from being loaded because it is missing integrity metadata.",
		body.Destination,
		body.BlockedUrl,
	)
}
