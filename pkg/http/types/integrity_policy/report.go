package integrity_policy

import "fmt"

type IntegrityViolationReportBody struct {
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
