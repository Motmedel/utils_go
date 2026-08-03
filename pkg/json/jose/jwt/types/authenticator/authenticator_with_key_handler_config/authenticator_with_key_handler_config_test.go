package authenticator_with_key_handler_config

import (
	"testing"
)

type stubMapValidator struct{}

func (stubMapValidator) Validate(_ map[string]any) error { return nil }

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	config := New()
	if config == nil {
		t.Fatal("New() returned nil")
	}
	if config.ClaimsValidator != nil {
		t.Errorf("default ClaimsValidator = %v, want nil", config.ClaimsValidator)
	}
	if config.HeaderValidator != nil {
		t.Errorf("default HeaderValidator = %v, want nil", config.HeaderValidator)
	}
}

func TestWithClaimsValidator(t *testing.T) {
	t.Parallel()

	claimsValidator := &stubMapValidator{}
	config := New(WithClaimsValidator(claimsValidator))
	if config.ClaimsValidator != claimsValidator {
		t.Errorf("ClaimsValidator = %v, want %v", config.ClaimsValidator, claimsValidator)
	}
}

func TestWithHeaderValidator(t *testing.T) {
	t.Parallel()

	headerValidator := &stubMapValidator{}
	config := New(WithHeaderValidator(headerValidator))
	if config.HeaderValidator != headerValidator {
		t.Errorf("HeaderValidator = %v, want %v", config.HeaderValidator, headerValidator)
	}
}

func TestNewAllOptions(t *testing.T) {
	t.Parallel()

	claimsValidator := &stubMapValidator{}
	headerValidator := &stubMapValidator{}

	config := New(
		WithClaimsValidator(claimsValidator),
		WithHeaderValidator(headerValidator),
	)

	if config.ClaimsValidator != claimsValidator {
		t.Errorf("ClaimsValidator = %v, want %v", config.ClaimsValidator, claimsValidator)
	}
	if config.HeaderValidator != headerValidator {
		t.Errorf("HeaderValidator = %v, want %v", config.HeaderValidator, headerValidator)
	}
}
