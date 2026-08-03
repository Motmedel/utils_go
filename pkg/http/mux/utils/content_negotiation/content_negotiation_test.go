package content_negotiation

import (
	"net/http"
	"testing"
)

func TestGetContentNegotiation(t *testing.T) {
	t.Parallel()

	t.Run("empty header is nil", func(t *testing.T) {
		t.Parallel()
		negotiation, err := GetContentNegotiation(http.Header{}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if negotiation != nil {
			t.Fatalf("expected nil, got %#v", negotiation)
		}
	})

	t.Run("parses accept headers", func(t *testing.T) {
		t.Parallel()
		header := http.Header{
			"Accept":          {"application/json"},
			"Accept-Encoding": {"gzip"},
			"Accept-Language": {"en-US"},
		}
		negotiation, err := GetContentNegotiation(header, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if negotiation == nil {
			t.Fatal("expected a content negotiation")
		}
		if negotiation.Accept == nil {
			t.Error("expected Accept to be parsed")
		}
		if negotiation.AcceptEncoding == nil {
			t.Error("expected AcceptEncoding to be parsed")
		}
		if negotiation.AcceptLanguage == nil {
			t.Error("expected AcceptLanguage to be parsed")
		}
	})
}
