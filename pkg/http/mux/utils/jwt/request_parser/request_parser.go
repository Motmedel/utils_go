package request_parser

import (
	"errors"
	"fmt"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	motmedelJwt "github.com/Motmedel/utils_go/pkg/jwt"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
	motmedelJwtToken "github.com/Motmedel/utils_go/pkg/jwt/types/token"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/validation_configuration"
	"net/http"
)

var (
	ErrEmptyName = errors.New("empty name")
)

type TokenWithRaw struct {
	motmedelJwtToken.Token
	Raw string
}

type RequestParser struct {
	Name              string
	SignatureVerifier motmedelCryptoInterfaces.NamedVerifier
	ClaimsValidator   validator.Validator[parsed_claims.ParsedClaims]
	HeaderValidator   validator.Validator[map[string]any]
}

func (parser *RequestParser) getToken(tokenString string) (*TokenWithRaw, *muxResponseError.ResponseError) {
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
	validationConfiguration := &validation_configuration.ValidationConfiguration{
		HeaderValidator:  parser.HeaderValidator,
		PayloadValidator: parser.ClaimsValidator,
	}

	token, err := motmedelJwt.ParseAndCheckWithConfiguration(
		tokenString,
		signatureVerifier,
		validationConfiguration,
	)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(
			fmt.Errorf("parse and validate with validator: %w", err),
			tokenString, signatureVerifier, validationConfiguration,
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

	return &TokenWithRaw{Token: *token, Raw: tokenString}, nil
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

func (parser *UrlRequestParser) Parse(request *http.Request) (*TokenWithRaw, *muxResponseError.ResponseError) {
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

func (parser *CookieRequestParser) Parse(request *http.Request) (*TokenWithRaw, *muxResponseError.ResponseError) {
	tokenString, responseError := parser.getTokenString(request)
	if responseError != nil {
		return nil, responseError
	}
	return parser.getToken(tokenString)
}
