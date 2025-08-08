package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrNilToken         = errors.New("nil token")
	ErrNilTokenHeader = errors.New("nil token header")
	ErrNilRawToken = errors.New("nil raw token")
	ErrEmptyTokenString = errors.New("empty token string")
	ErrNilRegisteredClaims = errors.New("nil registered claims")
	ErrAlgorithmMismatch = errors.New("algorithm mismatch")
	ErrMissingRequiredClaim = errors.New("missing required claim")
	ErrIssuerMismatch = errors.New("issuer mismatch")
	ErrAudienceMismatch = errors.New("audience mismatch")
	ErrSubjectMismatch = errors.New("subject mismatch")
	ErrExpExpired = errors.New("exp expired")
	ErrNbfBefore  = errors.New("nbf before")
	ErrIatBefore = errors.New("iat before")
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
