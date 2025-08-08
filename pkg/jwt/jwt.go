package jwt

import (
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelValidator "github.com/Motmedel/utils_go/pkg/interfaces/validator"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/raw_token"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
	jwtToken "github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/registered_claims_validator"
	"github.com/Motmedel/utils_go/pkg/utils"
)

func ParseAndValidateWithValidator(
	tokenString string,
	signatureVerifier motmedelCryptoInterfaces.NamedVerifier,
	claimsValidator motmedelValidator.Validator[parsed_claims.ParsedClaims],
) (*jwtToken.Token, error) {
	if utils.IsNil(signatureVerifier) {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	if utils.IsNil(claimsValidator) {
		return nil, motmedelErrors.NewWithTrace(motmedelValidator.ErrNilValidator)
	}

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

	tokenPayload := token.Payload
	parsedClaims, err := parsed_claims.FromMap(tokenPayload)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("%w: make parsed claims: %w", motmedelErrors.ErrParseError, err),
			tokenPayload,
		)
	}

	if err := claimsValidator.Validate(parsedClaims); err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("validator validate: %w", err), parsedClaims)
	}

	return token, nil
}

func ParseAndValidate(tokenString string, signatureVerifier motmedelCryptoInterfaces.NamedVerifier) (*jwtToken.Token, error) {
	var claimsValidator registered_claims_validator.RegisteredClaimsValidator
	return ParseAndValidateWithValidator(tokenString, signatureVerifier, &claimsValidator)
}
