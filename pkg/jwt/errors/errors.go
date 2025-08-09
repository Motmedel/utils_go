package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrNilToken                   = errors.New("nil token")
	ErrNilTokenHeader             = errors.New("nil token header")
	ErrNilRawToken                = errors.New("nil raw token")
	ErrEmptyTokenString           = errors.New("empty token string")
	ErrNilRegisteredClaims        = errors.New("nil registered claims")
	ErrAlgorithmMismatch          = errors.New("algorithm mismatch")
	ErrMissingRequiredField       = errors.New("missing required field")
	ErrIssuerMismatch             = errors.New("issuer mismatch")
	ErrAudienceMismatch           = errors.New("audience mismatch")
	ErrSubjectMismatch            = errors.New("subject mismatch")
	ErrIdMismatch                 = errors.New("id mismatch")
	ErrExpExpired                 = errors.New("exp expired")
	ErrNbfBefore                  = errors.New("nbf before")
	ErrIatBefore                  = errors.New("iat before")
	ErrNilValidationConfiguration = errors.New("nil validation configuration")
	ErrTypMismatch                = errors.New("typ mismatch")
	ErrAlgMismatch                = errors.New("alg mismatch")
	ErrNilNumericDate             = errors.New("nil numeric date")
)

type MissingRequiredFieldError struct {
	Name string
}

func (e *MissingRequiredFieldError) Error() string {
	if e == nil || e.Name == "" {
		return "missing required field"
	}

	return "missing required field: " + e.Name
}

func (e *MissingRequiredFieldError) Unwrap() error {
	return ErrMissingRequiredField
}
