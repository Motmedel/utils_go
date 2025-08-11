package jwk_validator

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/base_validator"
	"github.com/Motmedel/utils_go/pkg/maps"
	motmedelStrings "github.com/Motmedel/utils_go/pkg/strings"
	"github.com/Motmedel/utils_go/pkg/utils"
	"strings"
)

type JwkValidator struct {
	BaseValidator base_validator.BaseValidator
}

func (v *JwkValidator) Validate(token *token.Token) error {
	if baseValidator := v.BaseValidator; !utils.IsNil(baseValidator) {
		if err := baseValidator.Validate(token); err != nil {
			return fmt.Errorf("base validator validate: %w", err)
		}
	}

	if token == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilToken)
	}

	tokenHeader := token.Header
	if tokenHeader == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilTokenHeader)
	}

	alg, err := maps.MapGetConvert[string](tokenHeader, "alg")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert (alg): %w", err), tokenHeader)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}

		return wrappedErr
	}

	claims := token.Payload
	if claims == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilTokenPayload)
	}

	key, err := maps.MapGetConvert[map[string]any](claims, "key")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert (key): %w", err), claims)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}

		return wrappedErr
	}

	kty, err := maps.MapGetConvert[string](key, "kty")
	if err != nil {
		var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert (kty): %w", err), key)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
	}

	var expectedKty string
	if motmedelStrings.HasAnyPrefix(alg, "RS", "PS") {
		expectedKty = "RSA"
	} else if strings.HasPrefix(alg, "ES") {
		expectedKty = "EC"
	}

	if expectedKty != "" {
		if kty != expectedKty {
			return motmedelErrors.New(
				fmt.Errorf("%w: %w", motmedelErrors.ErrVerificationError, motmedelJwtErrors.ErrAlgKtyMismatch),
				alg, kty,
			)
		}

		if expectedKty == "EC" {
			if _, err := maps.MapGetConvert[string](key, "crv"); err != nil {
				return motmedelErrors.New(
					fmt.Errorf("%w: %w (crv)", motmedelErrors.ErrValidationError, err),
					key, "crv",
				)
			}
		}
	}

	return nil
}
