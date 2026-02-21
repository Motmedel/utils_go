package log_entry

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Motmedel/utils_go/pkg/cloud/gcp/types/log_entry/http_request"
)

type LogEntry struct {
	HttpRequest  *http_request.Request `json:"httpRequest,omitempty"`
	Trace        string                `json:"logging.googleapis.com/trace,omitempty"`
	TraceId      string                `json:"-"`
	SpanId       string                `json:"logging.googleapis.com/spanId,omitempty"`
	TraceSampled *bool                 `json:"logging.googleapis.com/trace_sampled,omitempty"`
}

// parseXCloudTraceContext parses the X-Cloud-Trace-Context header.
// Header format: TRACE_ID/SPAN_ID;o=TRACE_TRUE
// Example: 105445aa7843bc8bf206b120001000/123;o=1
// It returns traceID (32-char hex), spanID (16-hex-digit lowercase), and sampled flag.
func parseXCloudTraceContext(h string) (traceID string, spanIDHex string, sampled *bool) {
	if h == "" {
		return "", "", nil
	}

	// Split TRACE_ID from the rest
	parts := strings.SplitN(h, "/", 2)
	traceID = parts[0]

	if len(parts) > 1 {
		// parts[1] is like: SPAN_ID;o=1
		sub := strings.SplitN(parts[1], ";", 2)
		if len(sub) > 0 {
			// Convert decimal SPAN_ID to 16-hex-digit lowercase as required by Cloud Logging
			if n, err := strconv.ParseUint(sub[0], 10, 64); err == nil {
				spanIDHex = fmt.Sprintf("%016x", n)
			}
		}
		if len(sub) > 1 && strings.HasPrefix(sub[1], "o=") {
			v := sub[1] == "o=1"
			sampled = &v
		}
	}

	return
}

func ParseHttp(request *http.Request, response *http.Response) *LogEntry {
	if request == nil && response == nil {
		return nil
	}

	var httpRequest http_request.Request
	var traceId string
	var spanId string
	var sampled *bool

	if request != nil {
		httpRequest.RequestMethod = request.Method
		httpRequest.UserAgent = request.UserAgent()
		httpRequest.RemoteIp = request.RemoteAddr
		httpRequest.Referer = request.Referer()
		httpRequest.Protocol = fmt.Sprintf("HTTP/%d.%d", request.ProtoMajor, request.ProtoMinor)

		if requestHeader := request.Header; requestHeader != nil {
			traceId, spanId, sampled = parseXCloudTraceContext(requestHeader.Get("X-Cloud-Trace-Context"))
		}
	}

	if response != nil {
		httpRequest.Status = response.StatusCode
	}

	return &LogEntry{HttpRequest: &httpRequest, TraceId: traceId, SpanId: spanId, TraceSampled: sampled}
}
