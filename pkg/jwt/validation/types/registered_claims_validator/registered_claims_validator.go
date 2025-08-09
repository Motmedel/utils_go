package registered_claims_validator

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/claims_strings"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/numeric_date"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
	"github.com/Motmedel/utils_go/pkg/jwt/validation"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/setting"
	"github.com/Motmedel/utils_go/pkg/utils"
	"time"
)

type ExpectedRegisteredClaims struct {
	Issuer   string
	Subject  string
	Audience string
	Id       string
}

type RegisteredClaimsValidator struct {
	Settings map[string]setting.Setting
	Expected *ExpectedRegisteredClaims
}

func (validator *RegisteredClaimsValidator) Validate(parsedClaims parsed_claims.ParsedClaims) error {
	if parsedClaims == nil {
		return nil
	}

	expected := validator.Expected
	if expected == nil {
		expected = &ExpectedRegisteredClaims{}
	}

	var errs []error

	for key, value := range validator.Settings {
		if _, ok := parsedClaims[key]; value == setting.SettingRequired && !ok {
			errs = append(
				errs,
				&motmedelJwtErrors.MissingRequiredFieldError{Name: key},
			)
		}
	}

	for key, value := range parsedClaims {
		if claimSetting := validator.Settings[key]; claimSetting == setting.SettingSkip {
			continue
		}

		switch key {
		case "exp":
			expiresAt, err := utils.Convert[numeric_date.NumericDate](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if err := validation.ValidateExpiresAt(expiresAt.Time, time.Now()); err != nil {
				wrappedErr := motmedelErrors.New(
					fmt.Errorf("validate expires at: %w", err),
					expiresAt.Time,
				)
				if !errors.Is(err, motmedelErrors.ErrValidationError) {
					return wrappedErr
				}
				errs = append(errs, wrappedErr)
			}
		case "nbf":
			notBefore, err := utils.Convert[numeric_date.NumericDate](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if err := validation.ValidateNotBefore(notBefore.Time, time.Now()); err != nil {
				wrappedErr := motmedelErrors.New(
					fmt.Errorf("validate not before: %w", err),
					notBefore.Time,
				)
				if !errors.Is(err, motmedelErrors.ErrValidationError) {
					return wrappedErr
				}
				errs = append(errs, wrappedErr)
			}
		case "iat":
			issuedAt, err := utils.Convert[numeric_date.NumericDate](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if err := validation.ValidateIssuedAt(issuedAt.Time, time.Now()); err != nil {
				wrappedErr := motmedelErrors.New(
					fmt.Errorf("validate issued at: %w", err),
					issuedAt.Time,
				)
				if !errors.Is(err, motmedelErrors.ErrValidationError) {
					return wrappedErr
				}
				errs = append(errs, wrappedErr)
			}
		case "aud":
			if expectedAudience := expected.Audience; expectedAudience != "" {
				audiences, err := utils.Convert[claims_strings.ClaimStrings](value)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
				}

				var audienceMatched bool

				for _, audience := range audiences {
					if audience == expectedAudience {
						audienceMatched = true
						break
					}
				}

				if !audienceMatched {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrAudienceMismatch, expectedAudience, audiences),
					)
				}
			}
		case "iss":
			if expectedIssuer := expected.Issuer; expectedIssuer != "" {
				issuer, err := utils.Convert[string](value)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
				}

				if issuer != expectedIssuer {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrIssuerMismatch, expectedIssuer, issuer),
					)
				}
			}
		case "sub":
			if expectedSubject := expected.Subject; expectedSubject != "" {
				subject, err := utils.Convert[string](value)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
				}

				if subject != expectedSubject {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrSubjectMismatch, expectedSubject, subject),
					)
				}
			}
		case "jti":
			if expectedId := expected.Id; expectedId != "" {
				id, err := utils.Convert[string](value)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
				}

				if id != expectedId {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrIdMismatch, expectedId, id),
					)
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, errors.Join(errs...))
	}

	return nil
}
