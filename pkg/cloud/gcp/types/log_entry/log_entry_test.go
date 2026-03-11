package log_entry

import (
	"net/http"
	"testing"
)

func TestParseXCloudTraceContext_Full(t *testing.T) {
	traceID, spanID, sampled := parseXCloudTraceContext("105445aa7843bc8bf206b120001000/123;o=1")
	if traceID != "105445aa7843bc8bf206b120001000" {
		t.Errorf("expected trace id '105445aa7843bc8bf206b120001000', got %q", traceID)
	}
	if spanID != "000000000000007b" {
		t.Errorf("expected span id '000000000000007b', got %q", spanID)
	}
	if sampled == nil || !*sampled {
		t.Error("expected sampled=true")
	}
}

func TestParseXCloudTraceContext_NotSampled(t *testing.T) {
	traceID, spanID, sampled := parseXCloudTraceContext("abcdef/456;o=0")
	if traceID != "abcdef" {
		t.Errorf("expected trace id 'abcdef', got %q", traceID)
	}
	if spanID != "00000000000001c8" {
		t.Errorf("expected span id '00000000000001c8', got %q", spanID)
	}
	if sampled == nil || *sampled {
		t.Error("expected sampled=false")
	}
}

func TestParseXCloudTraceContext_TraceOnly(t *testing.T) {
	traceID, spanID, sampled := parseXCloudTraceContext("abc123")
	if traceID != "abc123" {
		t.Errorf("expected trace id 'abc123', got %q", traceID)
	}
	if spanID != "" {
		t.Errorf("expected empty span id, got %q", spanID)
	}
	if sampled != nil {
		t.Error("expected nil sampled")
	}
}

func TestParseXCloudTraceContext_TraceAndSpan(t *testing.T) {
	traceID, spanID, sampled := parseXCloudTraceContext("trace123/789")
	if traceID != "trace123" {
		t.Errorf("expected trace id 'trace123', got %q", traceID)
	}
	if spanID != "0000000000000315" {
		t.Errorf("expected span id '0000000000000315', got %q", spanID)
	}
	if sampled != nil {
		t.Error("expected nil sampled")
	}
}

func TestParseXCloudTraceContext_Empty(t *testing.T) {
	traceID, spanID, sampled := parseXCloudTraceContext("")
	if traceID != "" {
		t.Errorf("expected empty trace id, got %q", traceID)
	}
	if spanID != "" {
		t.Errorf("expected empty span id, got %q", spanID)
	}
	if sampled != nil {
		t.Error("expected nil sampled")
	}
}

func TestParseXCloudTraceContext_NonNumericSpan(t *testing.T) {
	traceID, spanID, sampled := parseXCloudTraceContext("trace/notanumber;o=1")
	if traceID != "trace" {
		t.Errorf("expected trace id 'trace', got %q", traceID)
	}
	if spanID != "" {
		t.Errorf("expected empty span id for non-numeric, got %q", spanID)
	}
	if sampled == nil || !*sampled {
		t.Error("expected sampled=true")
	}
}

func TestParseHttp_RequestAndResponse(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com/api", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.Header.Set("X-Cloud-Trace-Context", "traceid123/100;o=1")
	req.RemoteAddr = "192.168.1.1:1234"

	resp := &http.Response{StatusCode: 200}

	entry := ParseHttp(req, resp)
	if entry == nil {
		t.Fatal("expected non-nil log entry")
	}

	httpReq := entry.HttpRequest
	if httpReq.RequestMethod != http.MethodGet {
		t.Errorf("expected GET, got %q", httpReq.RequestMethod)
	}
	if httpReq.UserAgent != "TestAgent/1.0" {
		t.Errorf("expected user agent 'TestAgent/1.0', got %q", httpReq.UserAgent)
	}
	if httpReq.RemoteIp != "192.168.1.1:1234" {
		t.Errorf("expected remote ip '192.168.1.1:1234', got %q", httpReq.RemoteIp)
	}
	if httpReq.Protocol != "HTTP/1.1" {
		t.Errorf("expected 'HTTP/1.1', got %q", httpReq.Protocol)
	}
	if httpReq.Status != 200 {
		t.Errorf("expected status 200, got %d", httpReq.Status)
	}

	if entry.TraceId != "traceid123" {
		t.Errorf("expected trace id 'traceid123', got %q", entry.TraceId)
	}
	if entry.SpanId != "0000000000000064" {
		t.Errorf("expected span id '0000000000000064', got %q", entry.SpanId)
	}
	if entry.TraceSampled == nil || !*entry.TraceSampled {
		t.Error("expected sampled=true")
	}
}

func TestParseHttp_RequestOnly(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	entry := ParseHttp(req, nil)
	if entry == nil {
		t.Fatal("expected non-nil log entry")
	}
	if entry.HttpRequest.RequestMethod != http.MethodPost {
		t.Errorf("expected POST, got %q", entry.HttpRequest.RequestMethod)
	}
	if entry.HttpRequest.Status != 0 {
		t.Errorf("expected status 0, got %d", entry.HttpRequest.Status)
	}
}

func TestParseHttp_ResponseOnly(t *testing.T) {
	resp := &http.Response{StatusCode: 404}
	entry := ParseHttp(nil, resp)
	if entry == nil {
		t.Fatal("expected non-nil log entry")
	}
	if entry.HttpRequest.Status != 404 {
		t.Errorf("expected status 404, got %d", entry.HttpRequest.Status)
	}
	if entry.HttpRequest.RequestMethod != "" {
		t.Errorf("expected empty method, got %q", entry.HttpRequest.RequestMethod)
	}
}

func TestParseHttp_BothNil(t *testing.T) {
	entry := ParseHttp(nil, nil)
	if entry != nil {
		t.Error("expected nil for both nil request and response")
	}
}

func TestParseHttp_NoTraceHeader(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	entry := ParseHttp(req, nil)
	if entry.TraceId != "" {
		t.Errorf("expected empty trace id, got %q", entry.TraceId)
	}
	if entry.SpanId != "" {
		t.Errorf("expected empty span id, got %q", entry.SpanId)
	}
	if entry.TraceSampled != nil {
		t.Error("expected nil sampled")
	}
}
