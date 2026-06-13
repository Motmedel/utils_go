package gemini

import (
	"testing"
	"time"
)

const sample429Body = `{
  "error": {
    "code": 429,
    "message": "You exceeded your current quota ... Please retry in 43.6s.",
    "status": "RESOURCE_EXHAUSTED",
    "details": [
      {"@type": "type.googleapis.com/google.rpc.Help", "links": []},
      {"@type": "type.googleapis.com/google.rpc.QuotaFailure", "violations": []},
      {"@type": "type.googleapis.com/google.rpc.RetryInfo", "retryDelay": "43s"}
    ]
  }
}`

func TestRetryAfterFromResponse(t *testing.T) {
	delay := RetryAfterFromResponse(nil, []byte(sample429Body))
	if delay == nil {
		t.Fatal("expected a retry delay")
	}
	if *delay != 43*time.Second {
		t.Errorf("expected 43s, got %v", *delay)
	}
}

func TestRetryAfterFromResponse_NoRetryInfo(t *testing.T) {
	body := `{"error":{"code":400,"message":"bad request","details":[]}}`
	if delay := RetryAfterFromResponse(nil, []byte(body)); delay != nil {
		t.Errorf("expected nil for a body without RetryInfo, got %v", *delay)
	}
}

func TestRetryAfterFromResponse_EmptyAndInvalid(t *testing.T) {
	if delay := RetryAfterFromResponse(nil, nil); delay != nil {
		t.Errorf("expected nil for empty body, got %v", *delay)
	}
	if delay := RetryAfterFromResponse(nil, []byte("not json")); delay != nil {
		t.Errorf("expected nil for non-JSON body, got %v", *delay)
	}
}
