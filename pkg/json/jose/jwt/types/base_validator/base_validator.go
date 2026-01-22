package base_validator

import (
	"fmt"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/parsed_claims"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/tokenapi"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type BaseValidator struct {
	HeaderValidator  validator.Validator[map[string]any]
	PayloadValidator validator.Validator[parsed_claims.Claims]
}

func (v *BaseValidator) Validate(token tokenapi.Token) error {
	if token == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilToken)
	}

	tokenHeader := token.HeaderFields()
	if headerValidator := v.HeaderValidator; !utils.IsNil(headerValidator) {
		if err := headerValidator.Validate(tokenHeader); err != nil {
			return motmedelErrors.New(fmt.Errorf("header validator validate: %w", err), tokenHeader)
		}
	}

	if payloadValidator := v.PayloadValidator; !utils.IsNil(payloadValidator) {
		tokenPayload := token.Claims()
		parsedClaims, err := parsed_claims.New(tokenPayload)
		if err != nil {
			return motmedelErrors.New(
				fmt.Errorf("%w: make parsed claims: %w", motmedelErrors.ErrParseError, err),
				tokenPayload,
			)
		}

		if err := payloadValidator.Validate(parsedClaims); err != nil {
			return motmedelErrors.New(fmt.Errorf("payload validator validate: %w", err), parsedClaims)
		}
	}

	return nil
}
