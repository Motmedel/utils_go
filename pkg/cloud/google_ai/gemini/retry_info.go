package gemini

import (
	"encoding/json/v2"
	"net/http"
	"time"
)

// retryInfoTypeUrl identifies the google.rpc.RetryInfo entry in an API error's
// details. Google APIs convey the advised retry delay there rather than in a
// Retry-After header.
const retryInfoTypeUrl = "type.googleapis.com/google.rpc.RetryInfo"

type errorDetail struct {
	Type       string `json:"@type"`
	RetryDelay string `json:"retryDelay"`
}

type apiErrorEnvelope struct {
	Error struct {
		Details []errorDetail `json:"details"`
	} `json:"error"`
}

// RetryAfterFromResponse extracts the server-advised retry delay from a Gemini API
// error response body (a google.rpc.RetryInfo entry, e.g. "retryDelay": "43s"). It
// returns nil when the body carries no such hint. It satisfies
// retry_config.RetryAfterFunc, so callers can wire it into a retry configuration to
// honour the API's own rate-limit timing instead of relying on back-off alone.
func RetryAfterFromResponse(_ *http.Response, responseBody []byte) *time.Duration {
	if len(responseBody) == 0 {
		return nil
	}

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return nil
	}

	for _, detail := range envelope.Error.Details {
		if detail.Type != retryInfoTypeUrl || detail.RetryDelay == "" {
			continue
		}
		// The delay uses Go's duration syntax ("43s", "1.5s").
		if delay, err := time.ParseDuration(detail.RetryDelay); err == nil && delay > 0 {
			return &delay
		}
	}

	return nil
}
