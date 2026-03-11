package message_part

import (
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message/message_part/body"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message/message_part/header"
)

type MessagePart struct {
	PartId   string           `json:"partId,omitempty"`
	MimeType string           `json:"mimeType,omitempty"`
	Filename string           `json:"filename,omitempty"`
	Headers  []*header.Header `json:"headers,omitempty"`
	Body     *body.Body       `json:"body,omitempty"`
	Parts    []*MessagePart   `json:"parts,omitempty"`
}
