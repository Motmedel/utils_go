package base_validator

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
	jwtToken "github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type BaseValidator struct {
	HeaderValidator  validator.Validator[map[string]any]
	PayloadValidator validator.Validator[parsed_claims.ParsedClaims]
}

func (v *BaseValidator) Validate(token *jwtToken.Token) error {
	if token == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilToken)
	}

	tokenHeader := token.Header
	if headerValidator := v.HeaderValidator; !utils.IsNil(headerValidator) {
		if err := headerValidator.Validate(tokenHeader); err != nil {
			return motmedelErrors.New(fmt.Errorf("header validator validate: %w", err), tokenHeader)
		}
	}

	if payloadValidator := v.PayloadValidator; !utils.IsNil(payloadValidator) {
		tokenPayload := token.Payload
		parsedClaims, err := parsed_claims.FromMap(tokenPayload)
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
