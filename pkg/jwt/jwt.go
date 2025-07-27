package jwt

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"net/url"
)

func Validate(tokenString string, key []byte) (*jwt.RegisteredClaims, error) {
	if len(key) == 0 {
		return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrEmptySigningKey)
	}

	if tokenString == "" {
		return nil, fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, jwtErrors.ErrEmptyTokenString)
	}

	claims := &jwt.RegisteredClaims{}

	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token == nil {
				return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilToken)
			}

			// TODO: Is this check necessary?
			tokenMethod := token.Method
			if _, ok := tokenMethod.(*jwt.SigningMethodHMAC); !ok {
				return nil, motmedelErrors.NewWithTrace(
					fmt.Errorf("%w: %T", motmedelErrors.ErrConversionNotOk, tokenMethod),
					tokenMethod,
				)
			}

			return key, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: %w",
			motmedelErrors.ErrValidationError,
			fmt.Errorf("jwt parse with claims: %w", err),
		)
	}
	if parsedToken == nil {
		return nil, fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, jwtErrors.ErrNilToken)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, jwtErrors.ErrInvalidToken)
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
