package url_request_parser

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxUtilsJwt "github.com/Motmedel/utils_go/pkg/http/mux/utils/jwt"
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
	Options []jwt.ParserOption
}

func (parser *UrlRequestParser) Parse(request *http.Request) (*muxUtilsJwt.TokenClaims, *muxResponseError.ResponseError) {
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

	claims, err := motmedelJwt.Validate(tokenString, signingKey, parser.Options...)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("jwt validate: %w", err), tokenString, signingKey)
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

	return &muxUtilsJwt.TokenClaims{RegisteredClaims: claims, TokenString: tokenString}, nil
}
