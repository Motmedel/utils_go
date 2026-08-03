package query_extractor_config

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	if New().AllowAdditionalParameters {
		t.Error("expected AllowAdditionalParameters to default to false")
	}
	if !New(WithAllowAdditionalParameters(true)).AllowAdditionalParameters {
		t.Error("expected WithAllowAdditionalParameters(true) to set the field")
	}
}
