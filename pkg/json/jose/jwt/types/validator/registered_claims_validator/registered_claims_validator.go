package registered_claims_validator

import (
	"errors"
	"fmt"
	"time"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/interfaces/comparer"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims/claim_strings"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims/parsed_claims"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims/registered_claims/numeric_date"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/validator/setting"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type ExpectedRegisteredClaims struct {
	IssuerComparer   comparer.Comparer[string]
	SubjectComparer  comparer.Comparer[string]
	AudienceComparer comparer.Comparer[string]
	IdComparer       comparer.Comparer[string]
}

type RegisteredClaimsValidator struct {
	Settings map[string]setting.Setting
	Expected *ExpectedRegisteredClaims
}

func (validator *RegisteredClaimsValidator) Validate(parsedClaims parsed_claims.Claims) error {
	if parsedClaims == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilTokenPayload)
	}

	expected := validator.Expected
	if expected == nil {
		expected = &ExpectedRegisteredClaims{}
	}

	var errs []error

	for key, value := range validator.Settings {
		if _, ok := parsedClaims[key]; value == setting.Required && !ok {
			errs = append(
				errs,
				&motmedelJwtErrors.MissingRequiredFieldError{Name: key},
			)
		}
	}

	for key, value := range parsedClaims {
		if claimSetting := validator.Settings[key]; claimSetting == setting.Skip {
			continue
		}

		switch key {
		case "exp":
			expiresAt, err := utils.Convert[numeric_date.NumericDate](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if err := jwt.ValidateExpiresAt(expiresAt.Time, time.Now()); err != nil {
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

			if err := jwt.ValidateNotBefore(notBefore.Time, time.Now()); err != nil {
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

			if err := jwt.ValidateIssuedAt(issuedAt.Time, time.Now()); err != nil {
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
			audiences, err := utils.Convert[claim_strings.ClaimStrings](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if audienceComparer := expected.AudienceComparer; !utils.IsNil(audienceComparer) {
				var audienceMatched bool

				for _, audience := range audiences {
					ok, err := audienceComparer.Compare(audience)
					if err != nil {
						return motmedelErrors.New(fmt.Errorf("compare (%s): %w", key, err), audience)
					}
					if ok {
						audienceMatched = true
						break
					}
				}

				if !audienceMatched {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrAudienceMismatch, audienceComparer, audiences),
					)
				}
			}
		case "iss":
			issuer, err := utils.Convert[string](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if issuerComparer := expected.IssuerComparer; !utils.IsNil(issuerComparer) {
				ok, err := issuerComparer.Compare(issuer)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("compare (%s): %w", key, err), issuer)
				}
				if !ok {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrIssuerMismatch, issuerComparer, issuer),
					)
				}
			}
		case "sub":
			subject, err := utils.Convert[string](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if subjectComparer := expected.SubjectComparer; !utils.IsNil(subjectComparer) {
				ok, err := subjectComparer.Compare(subject)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("compare (%s): %w", key, err), subject)
				}
				if !ok {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrSubjectMismatch, subjectComparer, subject),
					)
				}
			}
		case "jti":
			id, err := utils.Convert[string](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if idComparer := expected.IdComparer; !utils.IsNil(idComparer) {
				ok, err := idComparer.Compare(id)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("compare (%s): %w", key, err), id)
				}
				if !ok {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrIdMismatch, idComparer, id),
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
