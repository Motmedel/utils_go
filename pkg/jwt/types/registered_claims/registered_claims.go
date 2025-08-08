package registered_claims

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/claims_strings"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/numeric_date"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type RegisteredClaims struct {
	// the `iss` (Issuer) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1
	Issuer string `json:"iss,omitempty"`

	// the `sub` (Subject) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.2
	Subject string `json:"sub,omitempty"`

	// the `aud` (Audience) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.3
	Audience claims_strings.ClaimStrings `json:"aud,omitempty"`

	// the `exp` (Expiration Time) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.4
	ExpiresAt *numeric_date.NumericDate `json:"exp,omitempty"`

	// the `nbf` (Not Before) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5
	NotBefore *numeric_date.NumericDate `json:"nbf,omitempty"`

	// the `iat` (Issued At) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.6
	IssuedAt *numeric_date.NumericDate `json:"iat,omitempty"`

	// the `jti` (JWT ID) claim. See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.7
	Id string `json:"jti,omitempty"`
}

func getNumericDate(value any) (*numeric_date.NumericDate, error) {
	switch typedValue := value.(type) {
	case *numeric_date.NumericDate:
		return typedValue, nil
	case numeric_date.NumericDate:
		return &typedValue, nil
	default:
		numericDate, err := numeric_date.Convert(typedValue)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("numeric date convert: %w", err), typedValue)
		}
		return numericDate, nil
	}
}

func FromMap(m map[string]any) (*RegisteredClaims, error) {
	if len(m) == 0 {
		return nil, nil
	}

	registeredClaims := &RegisteredClaims{}

	// iss
	if v, ok := m["iss"]; ok && v != nil {
		vs, err := utils.Convert[string](v)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("convert (iss): %w", err), v)
		}
		registeredClaims.Issuer = vs
	}

	// sub
	if v, ok := m["sub"]; ok && v != nil {
		vs, err := utils.Convert[string](v)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("convert (sub): %w", err), v)
		}
		registeredClaims.Subject = vs
	}

	// aud
	if v, ok := m["aud"]; ok && v != nil {
		aud, err := claims_strings.Convert(v)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("claims strings convert: %w", err), v)
		}
		registeredClaims.Audience = aud
	}

	// exp
	if v, ok := m["exp"]; ok {
		if v == nil {
			registeredClaims.ExpiresAt = nil
		} else {
			numericDate, err := getNumericDate(v)
			if err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("get numeric date (exp): %w", err), v)
			}
			registeredClaims.ExpiresAt = numericDate
		}
	}

	// nbf
	if v, ok := m["nbf"]; ok {
		if v == nil {
			registeredClaims.NotBefore = nil
		} else {
			numericDate, err := getNumericDate(v)
			if err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("get numeric date (nbf): %w", err), v)
			}
			registeredClaims.NotBefore = numericDate
		}
	}

	// iat
	if v, ok := m["iat"]; ok {
		if v == nil {
			registeredClaims.IssuedAt = nil
		} else {
			numericDate, err := getNumericDate(v)
			if err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("get numeric date (iat): %w", err), v)
			}
			registeredClaims.IssuedAt = numericDate
		}
	}

	// jti
	if v, ok := m["jti"]; ok {
		vs, err := utils.Convert[string](v)
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("convert (jti): %w", err), v)
		}
		registeredClaims.Id = vs
	}

	return registeredClaims, nil
}
