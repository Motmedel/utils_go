package jwk

import (
	"fmt"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/maps"
	motmedelStrings "github.com/Motmedel/utils_go/pkg/strings"
)

func Validate(keyMap map[string]any) error {
	if keyMap == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelErrors.ErrNilMap)
	}

	kty, err := maps.MapGetConvert[string](keyMap, "kty")
	if err != nil {
		wrappedErr := fmt.Errorf("map get convert (kty): %w", err)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
		return wrappedErr
	}

	alg, err := maps.MapGetConvert[string](keyMap, "alg")
	if err != nil {
		wrappedErr := fmt.Errorf("map get convert (alg): %w", err)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
			wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
		}
		return wrappedErr
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
			if _, err := maps.MapGetConvert[string](keyMap, "crv"); err != nil {
				return motmedelErrors.New(fmt.Errorf("%w: %w (crv)", motmedelErrors.ErrValidationError, err))
			}
		}
	}

	return nil
}
