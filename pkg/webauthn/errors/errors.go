package errors

import "errors"

var (
	ErrCredentialTypeMismatch          = errors.New("credential type mismatch")
	ErrCollectedClientDataTypeMismatch = errors.New("collected client data type mismatch")
	ErrChallengeMismatch               = errors.New("challenge mismatch")
	ErrOriginMismatch                  = errors.New("origin mismatch")
	ErrRpIdHashMismatch                = errors.New("rp id hash mismatch")
	ErrUserNotPresent                  = errors.New("user not present")
	ErrUserNotVerified                 = errors.New("user not verified")
	ErrUnexpectedSignatureCount        = errors.New("unexpected signature count")
	ErrPublicKeyAlgorithmMismatch      = errors.New("public key algorithm mismatch")
	ErrPublicKeyMismatch               = errors.New("public key mismatch")
	ErrAuthenticatorDataMismatch       = errors.New("authenticator data mismatch")
	ErrSignatureVerifyFailure          = errors.New("signature verify failure")
	ErrUnsupportedAttestationFormat    = errors.New("unsupported attestation format")
)
