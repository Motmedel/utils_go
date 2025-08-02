package jwt

import "github.com/golang-jwt/jwt/v5"

type TokenStringClaims interface {
	jwt.Claims
	SetTokenString(string)
}
