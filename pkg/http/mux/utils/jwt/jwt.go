package jwt

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	*jwt.RegisteredClaims
	TokenString string
}
