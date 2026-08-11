package message

import (
	"encoding/base64"
	"encoding/json/v2"
	"maps"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		payload    []byte
		attributes map[string]string
		wantData   string
	}{
		{
			name:       "non-empty payload with attributes",
			payload:    []byte("hello world"),
			attributes: map[string]string{"k": "v"},
			wantData:   base64.StdEncoding.EncodeToString([]byte("hello world")),
		},
		{
			name:       "nil attributes",
			payload:    []byte("data"),
			attributes: nil,
			wantData:   base64.StdEncoding.EncodeToString([]byte("data")),
		},
		{
			name:       "empty payload",
			payload:    []byte{},
			attributes: nil,
			wantData:   "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			msg := New(testCase.payload, testCase.attributes)
			if msg == nil {
				t.Fatal("New returned nil")
			}
			if msg.Data != testCase.wantData {
				t.Errorf("Data = %q, want %q", msg.Data, testCase.wantData)
			}
			if !maps.Equal(msg.Attributes, testCase.attributes) {
				t.Errorf("Attributes = %#v, want %#v", msg.Attributes, testCase.attributes)
			}
			if msg.MessageId != "" || msg.OrderingKey != "" || msg.PublishTime != "" {
				t.Errorf("expected MessageId/OrderingKey/PublishTime empty, got %#v", msg)
			}

			// Data must decode back to the original payload.
			decoded, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				t.Fatalf("base64 decode: %v", err)
			}
			if string(decoded) != string(testCase.payload) {
				t.Errorf("decoded Data = %q, want %q", decoded, testCase.payload)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	msg := New([]byte("hi"), nil)
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	// omitempty must drop the empty attributes/id/time fields.
	want := `{"data":"aGk="}`
	if string(data) != want {
		t.Errorf("marshaled = %s, want %s", data, want)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	raw := `{"data":"aGk=","attributes":{"a":"b"},"messageId":"123","orderingKey":"ok","publishTime":"2024-01-01T00:00:00Z"}`
	var msg Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if msg.Data != "aGk=" {
		t.Errorf("Data = %q, want %q", msg.Data, "aGk=")
	}
	if msg.MessageId != "123" {
		t.Errorf("MessageId = %q, want %q", msg.MessageId, "123")
	}
	if msg.OrderingKey != "ok" {
		t.Errorf("OrderingKey = %q, want %q", msg.OrderingKey, "ok")
	}
	if msg.PublishTime != "2024-01-01T00:00:00Z" {
		t.Errorf("PublishTime = %q, want %q", msg.PublishTime, "2024-01-01T00:00:00Z")
	}
	if !maps.Equal(msg.Attributes, map[string]string{"a": "b"}) {
		t.Errorf("Attributes = %#v", msg.Attributes)
	}
}
