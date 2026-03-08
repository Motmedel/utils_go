package integrity_policy

type IntegrityViolationReportBody struct {
	DocumentUrl string `json:"documentURL"`
	BlockedUrl  string `json:"blockedURL"`
	Destination string `json:"destination"`
	ReportOnly  bool   `json:"reportOnly"`
}
