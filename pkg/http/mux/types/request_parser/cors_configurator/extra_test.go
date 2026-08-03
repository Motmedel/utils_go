package cors_configurator

import (
	"net/http"
	"testing"
)

func TestParse_MalformedOriginWithRegisteredDomain(t *testing.T) {
	t.Parallel()

	configurator := &Configurator{RegisteredDomain: "example.com"}
	_, responseError := configurator.Parse(newRequestWithOrigin(t, "http://[::1"))
	if responseError == nil || responseError.ProblemDetail == nil || responseError.ProblemDetail.Status != http.StatusBadRequest {
		t.Fatalf("expected 400 for a malformed Origin, got %#v", responseError)
	}
}
