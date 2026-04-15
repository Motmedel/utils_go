package gmail

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/get_message_config"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/message"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/send_as"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/watch_request"
	"github.com/Motmedel/utils_go/pkg/cloud/gws/gmail/types/watch_response"
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
		json.UnmarshalRead(r.Body, &input)

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, &message.Message{
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

func TestWatch(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/watch") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input watch_request.WatchRequest
		json.UnmarshalRead(r.Body, &input)

		if input.TopicName != "projects/my-project/topics/my-topic" {
			t.Errorf("expected topic 'projects/my-project/topics/my-topic', got %q", input.TopicName)
		}

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, &watch_response.WatchResponse{
			HistoryId:  "12345",
			Expiration: "1431990098200",
		})
	})

	resp, err := client.Watch(context.Background(), "me", &watch_request.WatchRequest{
		TopicName: "projects/my-project/topics/my-topic",
		LabelIds:  []string{"INBOX"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HistoryId != "12345" {
		t.Errorf("expected history id '12345', got %q", resp.HistoryId)
	}
	if resp.Expiration != "1431990098200" {
		t.Errorf("expected expiration '1431990098200', got %q", resp.Expiration)
	}
}

func TestWatch_EmptyUserId(t *testing.T) {
	client := NewClient()
	_, err := client.Watch(context.Background(), "", &watch_request.WatchRequest{
		TopicName: "projects/my-project/topics/my-topic",
	})
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestWatch_NilRequest(t *testing.T) {
	client := NewClient()
	resp, err := client.Watch(context.Background(), "me", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != nil {
		t.Error("expected nil for nil request")
	}
}

func TestWatch_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Watch(ctx, "me", &watch_request.WatchRequest{
		TopicName: "projects/my-project/topics/my-topic",
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestWatchUrl(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	client := NewClientWithBaseUrl(u)
	got := client.watchUrl("user@example.com")
	expected := "http://localhost:8080/gmail/v1/users/user@example.com/watch"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
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

func TestCreateSendAs(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/settings/sendAs") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input send_as.SendAs
		json.UnmarshalRead(r.Body, &input)

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, &send_as.SendAs{
			SendAsEmail:        input.SendAsEmail,
			DisplayName:        input.DisplayName,
			TreatAsAlias:       true,
			VerificationStatus: "accepted",
		})
	})

	created, err := client.CreateSendAs(context.Background(), "me", &send_as.SendAs{
		SendAsEmail: "support@example.com",
		DisplayName: "Support",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.SendAsEmail != "support@example.com" {
		t.Errorf("expected email 'support@example.com', got %q", created.SendAsEmail)
	}
	if created.VerificationStatus != "accepted" {
		t.Errorf("expected status 'accepted', got %q", created.VerificationStatus)
	}
}

func TestCreateSendAs_EmptyUserId(t *testing.T) {
	client := NewClient()
	_, err := client.CreateSendAs(context.Background(), "", &send_as.SendAs{SendAsEmail: "a@b.com"})
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestCreateSendAs_NilSendAs(t *testing.T) {
	client := NewClient()
	result, err := client.CreateSendAs(context.Background(), "me", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for nil send-as")
	}
}

func TestCreateSendAs_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.CreateSendAs(ctx, "me", &send_as.SendAs{SendAsEmail: "a@b.com"})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestGetSendAs(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/settings/sendAs/support@example.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.MarshalWrite(
			w,
			&send_as.SendAs{
				SendAsEmail:        "support@example.com",
				DisplayName:        "Support",
				VerificationStatus: "accepted",
			},
		)
		if err != nil {
			t.Fatalf("json marshal write: %v", err)
		}
	})

	s, err := client.GetSendAs(context.Background(), "me", "support@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SendAsEmail != "support@example.com" {
		t.Errorf("expected email 'support@example.com', got %q", s.SendAsEmail)
	}
}

func TestGetSendAs_EmptyUserId(t *testing.T) {
	client := NewClient()
	_, err := client.GetSendAs(context.Background(), "", "a@b.com")
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestGetSendAs_EmptySendAsEmail(t *testing.T) {
	client := NewClient()
	_, err := client.GetSendAs(context.Background(), "me", "")
	if err == nil {
		t.Fatal("expected error for empty send-as email")
	}
}

func TestUpdateSendAs(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/settings/sendAs/support@example.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var input send_as.SendAs
		json.UnmarshalRead(r.Body, &input)

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, &send_as.SendAs{
			SendAsEmail: "support@example.com",
			DisplayName: input.DisplayName,
		})
	})

	updated, err := client.UpdateSendAs(context.Background(), "me", "support@example.com", &send_as.SendAs{
		DisplayName: "New Support Name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.DisplayName != "New Support Name" {
		t.Errorf("expected display name 'New Support Name', got %q", updated.DisplayName)
	}
}

func TestUpdateSendAs_EmptySendAsEmail(t *testing.T) {
	client := NewClient()
	_, err := client.UpdateSendAs(context.Background(), "me", "", &send_as.SendAs{})
	if err == nil {
		t.Fatal("expected error for empty send-as email")
	}
}

func TestUpdateSendAs_NilSendAs(t *testing.T) {
	client := NewClient()
	result, err := client.UpdateSendAs(context.Background(), "me", "a@b.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for nil send-as")
	}
}

func TestDeleteSendAs(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/settings/sendAs/support@example.com") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.DeleteSendAs(context.Background(), "me", "support@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSendAs_EmptyUserId(t *testing.T) {
	client := NewClient()
	err := client.DeleteSendAs(context.Background(), "", "a@b.com")
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestDeleteSendAs_EmptySendAsEmail(t *testing.T) {
	client := NewClient()
	err := client.DeleteSendAs(context.Background(), "me", "")
	if err == nil {
		t.Fatal("expected error for empty send-as email")
	}
}

func TestListMessages(t *testing.T) {
	callCount := 0
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/messages") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "in:inbox" {
			t.Errorf("expected q=in:inbox, got %q", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			json.MarshalWrite(w, map[string]any{
				"messages": []map[string]string{
					{"id": "msg-1", "threadId": "thread-1"},
					{"id": "msg-2", "threadId": "thread-2"},
				},
				"nextPageToken":      "token-abc",
				"resultSizeEstimate": 3,
			})
		} else {
			json.MarshalWrite(w, map[string]any{
				"messages": []map[string]string{
					{"id": "msg-3", "threadId": "thread-3"},
				},
				"resultSizeEstimate": 3,
			})
		}
	})

	messages, err := client.ListMessages(context.Background(), "me", "in:inbox")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}
	if messages[0].Id != "msg-1" {
		t.Errorf("expected first message id 'msg-1', got %q", messages[0].Id)
	}
	if messages[2].Id != "msg-3" {
		t.Errorf("expected third message id 'msg-3', got %q", messages[2].Id)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestListMessages_EmptyUserId(t *testing.T) {
	client := NewClient()
	_, err := client.ListMessages(context.Background(), "", "in:inbox")
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestListMessages_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.ListMessages(ctx, "me", "in:inbox")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestListMessages_EmptyQuery(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "" {
			t.Errorf("expected no q parameter, got %q", r.URL.Query().Get("q"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, map[string]any{
			"messages": []map[string]string{
				{"id": "msg-1", "threadId": "thread-1"},
			},
		})
	})

	messages, err := client.ListMessages(context.Background(), "me", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
}

func TestGetMessage(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/me/messages/msg-123") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, &message.Message{
			Id:           "msg-123",
			ThreadId:     "thread-456",
			LabelIds:     []string{"INBOX"},
			Snippet:      "Hello world",
			InternalDate: "1234567890000",
			SizeEstimate: 1024,
		})
	})

	msg, err := client.GetMessage(context.Background(), "me", "msg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Id != "msg-123" {
		t.Errorf("expected id 'msg-123', got %q", msg.Id)
	}
	if msg.ThreadId != "thread-456" {
		t.Errorf("expected thread id 'thread-456', got %q", msg.ThreadId)
	}
	if msg.Snippet != "Hello world" {
		t.Errorf("expected snippet 'Hello world', got %q", msg.Snippet)
	}
}

func TestGetMessage_WithFormatAndMetadataHeaders(t *testing.T) {
	client := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "metadata" {
			t.Errorf("expected format=metadata, got %q", r.URL.Query().Get("format"))
		}
		headers := r.URL.Query()["metadataHeaders"]
		if len(headers) != 2 || headers[0] != "Subject" || headers[1] != "From" {
			t.Errorf("expected metadataHeaders=[Subject, From], got %v", headers)
		}

		w.Header().Set("Content-Type", "application/json")
		json.MarshalWrite(w, &message.Message{
			Id: "msg-123",
		})
	})

	msg, err := client.GetMessage(
		context.Background(), "me", "msg-123",
		get_message_config.WithFormat(get_message_config.FormatMetadata),
		get_message_config.WithMetadataHeaders("Subject", "From"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Id != "msg-123" {
		t.Errorf("expected id 'msg-123', got %q", msg.Id)
	}
}

func TestGetMessage_EmptyUserId(t *testing.T) {
	client := NewClient()
	_, err := client.GetMessage(context.Background(), "", "msg-123")
	if err == nil {
		t.Fatal("expected error for empty user id")
	}
}

func TestGetMessage_EmptyMessageId(t *testing.T) {
	client := NewClient()
	_, err := client.GetMessage(context.Background(), "me", "")
	if err == nil {
		t.Fatal("expected error for empty message id")
	}
}

func TestGetMessage_CancelledContext(t *testing.T) {
	client := NewClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.GetMessage(ctx, "me", "msg-123")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestMessagesUrl(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	client := NewClientWithBaseUrl(u)

	got := client.messagesUrl("user@example.com", "")
	expected := "http://localhost:8080/gmail/v1/users/user@example.com/messages"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	got = client.messagesUrl("user@example.com", "msg-123")
	expected = "http://localhost:8080/gmail/v1/users/user@example.com/messages/msg-123"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSendAsUrl(t *testing.T) {
	u, _ := url.Parse("http://localhost:8080")
	client := NewClientWithBaseUrl(u)

	got := client.sendAsUrl("user@example.com", "")
	expected := "http://localhost:8080/gmail/v1/users/user@example.com/settings/sendAs"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}

	got = client.sendAsUrl("user@example.com", "alias@example.com")
	expected = "http://localhost:8080/gmail/v1/users/user@example.com/settings/sendAs/alias@example.com"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
