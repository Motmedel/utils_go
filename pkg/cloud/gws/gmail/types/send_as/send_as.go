package send_as

type SendAs struct {
	SendAsEmail        string `json:"sendAsEmail,omitzero"`
	DisplayName        string `json:"displayName,omitzero"`
	ReplyToAddress     string `json:"replyToAddress,omitzero"`
	Signature          string `json:"signature,omitzero"`
	IsPrimary          bool   `json:"isPrimary,omitzero"`
	IsDefault          bool   `json:"isDefault,omitzero"`
	TreatAsAlias       bool   `json:"treatAsAlias,omitzero"`
	VerificationStatus string `json:"verificationStatus,omitzero"`
}
