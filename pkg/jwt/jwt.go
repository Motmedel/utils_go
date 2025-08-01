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

func Validate[T jwt.Claims](tokenString string, key []byte, options ...jwt.ParserOption) (T, error) {
	var claims T

	if len(key) == 0 {
		return claims, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptySigningKey)
	}

	if tokenString == "" {
		return claims, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptyTokenString)
	}

	var parserOptions []jwt.ParserOption
	if len(options) > 0 {
		parserOptions = append(parserOptions, jwt.WithoutClaimsValidation())
	}

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
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
		parserOptions...,
	)
	if err != nil {
		return claims, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: jwt parse with claims: %w", motmedelErrors.ErrValidationError, err),
			parserOptions,
		)
	}
	if parsedToken == nil {
		return claims, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
	}

	if len(options) != 0 {
		validator := jwt.NewValidator(options...)
		if validator == nil {
			return claims, motmedelErrors.NewWithTrace(jwtErrors.ErrNilValidator)
		}

		if err := validator.Validate(claims); err != nil {
			return claims, fmt.Errorf("%w: jwt validator validate: %w", motmedelErrors.ErrValidationError, err)
		}
	}

	return claims, nil
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
