package jwt

import (
	"fmt"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/raw_token"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
	jwtToken "github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/registered_claims_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/validation_configuration"
	"github.com/Motmedel/utils_go/pkg/utils"
)

func ParseAndCheckWithConfiguration(
	tokenString string,
	signatureVerifier motmedelCryptoInterfaces.NamedVerifier,
	validationConfig *validation_configuration.ValidationConfiguration,
) (*jwtToken.Token, error) {
	if tokenString == "" {
		return nil, nil
	}

	rawToken, err := raw_token.New(tokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: raw token new: %w", motmedelErrors.ErrParseError, err)
	}
	if rawToken == nil {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilRawToken)
	}

	token, err := jwtToken.FromRawToken(rawToken)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("%w: token from raw token: %w", motmedelErrors.ErrParseError, err),
			rawToken,
		)
	}

	if !utils.IsNil(signatureVerifier) {
		if token == nil {
			return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
		}

		tokenHeader := token.Header
		if tokenHeader == nil {
			return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilTokenHeader)
		}

		tokenHeaderAlgorithm, _ := tokenHeader["alg"].(string)
		verifierMethodName := signatureVerifier.GetName()
		if tokenHeaderAlgorithm != verifierMethodName {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, jwtErrors.ErrAlgorithmMismatch),
				jwtErrors.ErrAlgorithmMismatch, tokenHeaderAlgorithm, verifierMethodName,
			)
		}

		if err := rawToken.Verify(signatureVerifier); err != nil {
			return nil, motmedelErrors.New(
				fmt.Errorf("%w: raw token verify: %w", motmedelErrors.ErrVerificationError, err),
				rawToken,
			)
		}
	}

	if config := validationConfig; config != nil {
		if token == nil {
			return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
		}

		tokenHeader := token.Header
		if tokenHeader == nil {
			return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilTokenHeader)
		}

		if validator := config.HeaderValidator; !utils.IsNil(validator) {
			if err := validator.Validate(tokenHeader); err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("header validator validate: %w", err), tokenHeader)
			}
		}

		if validator := config.PayloadValidator; !utils.IsNil(validator) {
			tokenPayload := token.Payload
			parsedClaims, err := parsed_claims.FromMap(tokenPayload)
			if err != nil {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w: make parsed claims: %w", motmedelErrors.ErrParseError, err),
					tokenPayload,
				)
			}

			if err := validator.Validate(parsedClaims); err != nil {
				return nil, motmedelErrors.New(fmt.Errorf("payload validator validate: %w", err), parsedClaims)
			}
		}
	}

	return token, nil
}

func ParseAndCheck(tokenString string, signatureVerifier motmedelCryptoInterfaces.NamedVerifier) (*jwtToken.Token, error) {
	return ParseAndCheckWithConfiguration(
		tokenString,
		signatureVerifier,
		&validation_configuration.ValidationConfiguration{
			PayloadValidator: &registered_claims_validator.RegisteredClaimsValidator{},
		},
	)
}
