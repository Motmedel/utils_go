package message

import "encoding/base64"

// Message mirrors the Pub/Sub REST PubsubMessage resource. Note that Data is the
// base64-encoded message payload, as required by the JSON API.
type Message struct {
	Data        string            `json:"data,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	MessageId   string            `json:"messageId,omitempty"`
	OrderingKey string            `json:"orderingKey,omitempty"`
	PublishTime string            `json:"publishTime,omitempty"`
}

// New constructs a Message from a raw (unencoded) payload, base64-encoding it as the
// REST API expects. Attributes may be nil.
func New(payload []byte, attributes map[string]string) *Message {
	return &Message{
		Data:       base64.StdEncoding.EncodeToString(payload),
		Attributes: attributes,
	}
}
