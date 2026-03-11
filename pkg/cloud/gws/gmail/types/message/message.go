package message

import (
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message/message_part"
)

type Message struct {
	Id           string   `json:"id,omitempty"`
	ThreadId     string   `json:"threadId,omitempty"`
	LabelIds     []string `json:"labelIds,omitempty"`
	Snippet      string   `json:"snippet,omitempty"`
	HistoryId    string   `json:"historyId,omitempty"`
	InternalDate string   `json:"internalDate,omitempty"`
	SizeEstimate int      `json:"sizeEstimate,omitempty"`
	Raw          string   `json:"raw,omitempty"`

	Payload *message_part.MessagePart `json:"payload,omitempty"`
}
