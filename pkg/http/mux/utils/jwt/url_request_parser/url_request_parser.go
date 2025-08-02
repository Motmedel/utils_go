package url_request_parser

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelJwt "github.com/Motmedel/utils_go/pkg/jwt"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

var (
	ErrEmptyParameterName = errors.New("empty parameter name")
)

type UrlRequestParser struct {
	ParameterName string
	SigningKey    []byte
	Options       []jwt.ParserOption
	Validator     func(jwt.Claims) error
}

func (parser *UrlRequestParser) Parse(request *http.Request) (*jwt.Token, *muxResponseError.ResponseError) {
	if request == nil {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestUrl := request.URL
	if requestUrl == nil {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestUrl),
		}
	}

	parameterName := parser.ParameterName
	if parameterName == "" {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(ErrEmptyParameterName),
		}
	}

	requestUrlQuery := requestUrl.Query()

	if !requestUrlQuery.Has(parameterName) {
		return nil, &muxResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnauthorized,
				"Missing token query parameter.",
				nil,
			),
		}
	}

	tokenString := requestUrlQuery.Get(parameterName)
	if tokenString == "" {
		return nil, &muxResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnauthorized,
				"Empty token query parameter.",
				nil,
			),
		}
	}

	signingKey := parser.SigningKey
	if len(signingKey) == 0 {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrEmptySigningKey),
		}
	}

	var token *jwt.Token
	var funcName string
	var err error

	if validator := parser.Validator; validator != nil {
		token, err = motmedelJwt.ValidateWithValidator(tokenString, signingKey, validator)
		funcName = "validate with validator"
	} else {
		token, err = motmedelJwt.Validate(tokenString, signingKey, parser.Options...)
		funcName = "validate"
	}

	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(
			fmt.Errorf("%s: %w", funcName, err),
			tokenString, signingKey,
		)
		if errors.Is(err, motmedelErrors.ErrValidationError) {
			return nil, &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnauthorized,
					"Invalid token query parameter.",
					nil,
				),
			}
		}
		return nil, &muxResponseError.ResponseError{ServerError: wrappedErr}
	}

	return token, nil
}
