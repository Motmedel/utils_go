package body

type Body struct {
	AttachmentId string `json:"attachmentId,omitzero"`
	Size         int    `json:"size,omitzero"`
	Data         string `json:"data,omitzero"`
}
