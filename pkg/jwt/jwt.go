package jwt

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/golang-jwt/jwt/v5"
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
