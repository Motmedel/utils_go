package request_parser

import (
	"errors"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	motmedelJwt "github.com/Motmedel/utils_go/pkg/jwt"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
	motmedelJwtToken "github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/registered_claims_validator"
	"github.com/Motmedel/utils_go/pkg/utils"
	"net/http"
)

var (
	ErrEmptyName = errors.New("empty name")
)

type RequestParser struct {
	Name              string
	SignatureVerifier motmedelCryptoInterfaces.NamedVerifier
	ClaimsValidator   validator.Validator[parsed_claims.ParsedClaims]
}

func (parser *RequestParser) getToken(tokenString string) (*motmedelJwtToken.Token, *muxResponseError.ResponseError) {
	if tokenString == "" {
		return nil, &muxResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnauthorized,
				"Empty token.",
				nil,
			),
		}
	}

	signatureVerifier := parser.SignatureVerifier
	if utils.IsNil(signatureVerifier) {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier),
		}
	}

	claimsValidator := parser.ClaimsValidator
	if utils.IsNil(claimsValidator) {
		claimsValidator = &registered_claims_validator.RegisteredClaimsValidator{}
	}

	token, err := motmedelJwt.ParseAndValidateWithValidator(tokenString, signatureVerifier, claimsValidator)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(
			fmt.Errorf("parse and validate with validator: %w", err),
			tokenString, signatureVerifier, claimsValidator,
		)
		if motmedelErrors.IsAny(err, motmedelErrors.ErrValidationError, motmedelErrors.ErrVerificationError, motmedelErrors.ErrParseError) {
			return nil, &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnauthorized,
					"Invalid token.",
					nil,
				),
			}
		}
		return nil, &muxResponseError.ResponseError{ServerError: wrappedErr}
	}

	return token, nil
}

type UrlRequestParser struct {
	RequestParser
}

func (parser *UrlRequestParser) getTokenString(request *http.Request) (string, *muxResponseError.ResponseError) {
	if request == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	name := parser.Name
	if name == "" {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(ErrEmptyName),
		}
	}

	requestUrl := request.URL
	if requestUrl == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestUrl),
		}
	}

	requestUrlQuery := requestUrl.Query()

	if !requestUrlQuery.Has(name) {
		return "", &muxResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnauthorized,
				"Missing token.",
				nil,
			),
		}
	}

	return requestUrlQuery.Get(name), nil
}

func (parser *UrlRequestParser) Parse(request *http.Request) (*motmedelJwtToken.Token, *muxResponseError.ResponseError) {
	tokenString, responseError := parser.getTokenString(request)
	if responseError != nil {
		return nil, responseError
	}
	return parser.getToken(tokenString)
}

type CookieRequestParser struct {
	RequestParser
}

func (parser *CookieRequestParser) getTokenString(request *http.Request) (string, *muxResponseError.ResponseError) {
	if request == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	name := parser.Name
	if name == "" {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(ErrEmptyName),
		}
	}

	tokenCookie, err := request.Cookie(name)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("request cookie: %w", err), name)
		if errors.Is(err, http.ErrNoCookie) {
			return "", &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnauthorized,
					"Missing token cookie.",
					nil,
				),
			}
		}
		return "", &muxResponseError.ResponseError{ServerError: wrappedErr}
	}
	if tokenCookie == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilCookie),
		}
	}

	return tokenCookie.Value, nil
}

func (parser *CookieRequestParser) Parse(request *http.Request) (*motmedelJwtToken.Token, *muxResponseError.ResponseError) {
	tokenString, responseError := parser.getTokenString(request)
	if responseError != nil {
		return nil, responseError
	}
	return parser.getToken(tokenString)
}
