package jwt

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"net/url"
)

var (
	BadRequestValidationErrors = []error{
		jwt.ErrTokenSignatureInvalid,
		jwt.ErrTokenExpired,
		jwt.ErrTokenUsedBeforeIssued,
		jwt.ErrTokenNotValidYet,
		jwt.ErrTokenInvalidAudience,
		jwt.ErrTokenInvalidIssuer,
		jwt.ErrTokenInvalidSubject,
	}
)

func ValidateWithValidator(tokenString string, key []byte, validator func(claims jwt.Claims) error) (*jwt.Token, error) {
	if len(key) == 0 {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptySigningKey)
	}

	if validator == nil {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilValidator)
	}

	if tokenString == "" {
		return nil, nil
	}

	token, err := Validate(tokenString, key, jwt.WithoutClaimsValidation())
	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}
	if token == nil {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
	}

	if err := validator(token.Claims); err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: validator: %w", motmedelErrors.ErrValidationError, err),
		)
	}

	return token, nil
}

func Validate(tokenString string, key []byte, options ...jwt.ParserOption) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptyTokenString)
	}

	if len(key) == 0 {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptySigningKey)
	}

	var claims jwt.RegisteredClaims

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token == nil {
				return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
			}

			tokenMethod := token.Method
			if _, ok := tokenMethod.(*jwt.SigningMethodHMAC); !ok {
				return nil, motmedelErrors.NewWithTrace(
					fmt.Errorf("%w: %T", motmedelErrors.ErrConversionNotOk, tokenMethod),
					tokenMethod,
				)
			}

			return key, nil
		},
		options...,
	)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: jwt parse with claims: %w", motmedelErrors.ErrValidationError, err),
		)
	}

	return parsedToken, nil
}

func MakeSignedUrlWithMethod(
	baseUrl url.URL,
	parameterName string,
	signingKey []byte,
	claims jwt.MapClaims,
	method jwt.SigningMethod,
) (*url.URL, error) {
	if len(signingKey) == 0 {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptySigningKey)
	}

	if parameterName == "" {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptyParameterName)
	}

	if utils.IsNil(method) {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilMethod)
	}

	token := jwt.NewWithClaims(method, claims)
	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("token signed string: %w", err), token)
	}

	registerTokenQuery := baseUrl.Query()
	registerTokenQuery.Set("token", signedToken)
	baseUrl.RawQuery = registerTokenQuery.Encode()

	return &baseUrl, nil
}

func MakeSignedUrl(baseUrl url.URL, parameterName string, signingKey []byte, claims jwt.MapClaims) (*url.URL, error) {
	if len(signingKey) == 0 {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptySigningKey)
	}

	if parameterName == "" {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptyParameterName)
	}

	signedUrl, err := MakeSignedUrlWithMethod(baseUrl, parameterName, signingKey, claims, jwt.SigningMethodHS256)
	if err != nil {
		return nil, fmt.Errorf("make signed url with method: %w", err)
	}

	return signedUrl, nil
}
