package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrNilToken         = errors.New("nil token")
	ErrEmptyTokenString = errors.New("empty token string")
	ErrEmptySigningKey  = errors.New("empty signing key")
	ErrNilMethod = errors.New("nil method")
	ErrEmptyParameterName = errors.New("empty parameter name")
	ErrNilValidator = errors.New("nil validator")
	ErrNilClaims = errors.New("nil claims")
	ErrNilTokenHeader = errors.New("nil token header")
	ErrNilRawToken = errors.New("nil raw token")
	ErrAlgorithmMismatch = errors.New("algorithm mismatch")
	ErrUnexpectedType = errors.New("unexpected type")
	ErrMissingRequiredClaim = errors.New("missing required claim")
	ErrIssuerMismatch = errors.New("issuer mismatch")
	ErrAudienceMismatch = errors.New("audience mismatch")
	ErrSubjectMismatch = errors.New("subject mismatch")
)

type MissingRequiredClaimError struct {
	Name string
}

func (e *MissingRequiredClaimError) Error() string {
	if e == nil || e.Name == "" {
		return "missing required claim"
	}

	return "missing required claim: " + e.Name
}

func (e *MissingRequiredClaimError) Unwrap() error {
	return ErrMissingRequiredClaim
}
