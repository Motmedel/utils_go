package authenticator_config

import (
	"testing"
)

type stubVerifier struct{}

func (stubVerifier) Verify(_ []byte, _ []byte) error { return nil }

func (stubVerifier) GetName() string { return "stub" }

type stubMapValidator struct{}

func (stubMapValidator) Validate(_ map[string]any) error { return nil }

func TestNewDefaults(t *testing.T) {
	t.Parallel()

	config := New()
	if config == nil {
		t.Fatal("New() returned nil")
	}
	if config.SignatureVerifier != nil {
		t.Errorf("default SignatureVerifier = %v, want nil", config.SignatureVerifier)
	}
	if config.ClaimsValidator != nil {
		t.Errorf("default ClaimsValidator = %v, want nil", config.ClaimsValidator)
	}
	if config.HeaderValidator != nil {
		t.Errorf("default HeaderValidator = %v, want nil", config.HeaderValidator)
	}
}

func TestWithSignatureVerifier(t *testing.T) {
	t.Parallel()

	verifier := &stubVerifier{}
	config := New(WithSignatureVerifier(verifier))
	if config.SignatureVerifier != verifier {
		t.Errorf("SignatureVerifier = %v, want %v", config.SignatureVerifier, verifier)
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

	verifier := &stubVerifier{}
	claimsValidator := &stubMapValidator{}
	headerValidator := &stubMapValidator{}

	config := New(
		WithSignatureVerifier(verifier),
		WithClaimsValidator(claimsValidator),
		WithHeaderValidator(headerValidator),
	)

	if config.SignatureVerifier != verifier {
		t.Errorf("SignatureVerifier = %v, want %v", config.SignatureVerifier, verifier)
	}
	if config.ClaimsValidator != claimsValidator {
		t.Errorf("ClaimsValidator = %v, want %v", config.ClaimsValidator, claimsValidator)
	}
	if config.HeaderValidator != headerValidator {
		t.Errorf("HeaderValidator = %v, want %v", config.HeaderValidator, headerValidator)
	}
}
