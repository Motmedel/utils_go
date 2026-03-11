package gmail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message"
)

func testServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return NewClientWithBaseUrl(u)
}

func TestSend(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/messages/send") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input message.Message
		json.NewDecoder(r.Body).Decode(&input)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&message.Message{
			Id:       "msg-123",
			ThreadId: "thread-456",
			LabelIds: []string{"SENT"},
		})
	})

	msg, err := client.Send(context.Background(), "me", &message.Message{
		Raw: "dGVzdA==",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Id != "msg-123" {
		t.Errorf("expected id 'msg-123', got %q", msg.Id)
	}
	if msg.ThreadId != "thread-456" {
		t.Errorf("expected thread id 'thread-456', got %q", msg.ThreadId)
	}
}

func TestSend_EmptyUserId(t *testing.T) {
	client := NewClient()
	_, err := client.Send(context.Background(), "", &message.Message{Raw: "dGVzdA=="})
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestSend_NilMessage(t *testing.T) {
	client := NewClient()
	msg, err := client.Send(context.Background(), "me", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Error("expected nil for nil message")
	}
}

func TestSend_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Send(ctx, "me", &message.Message{Raw: "dGVzdA=="})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestSendUrl(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	client := NewClientWithBaseUrl(u)
	got := client.sendUrl("user@example.com")
	expected := "http://localhost:8080/gmail/v1/users/user@example.com/messages/send"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
