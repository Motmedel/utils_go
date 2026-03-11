package body

type Body struct {
	AttachmentId string `json:"attachmentId,omitempty"`
	Size         int    `json:"size,omitempty"`
	Data         string `json:"data,omitempty"`
}
