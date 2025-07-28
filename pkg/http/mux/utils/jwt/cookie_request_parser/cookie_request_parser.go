package cookie_request_parser

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
	"net/http"
)

var (
	ErrEmptyCookieName = errors.New("empty cookie name")
)

type CookieRequestParser struct {
	CookieName string
	SigningKey []byte
}

func (c *CookieRequestParser) Parse(request *http.Request) (*muxUtilsJwt.TokenClaims, *muxResponseError.ResponseError) {
	if request == nil {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	cookieName := c.CookieName
	if cookieName == "" {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(ErrEmptyCookieName),
		}
	}

	// TODO: Not sure if `WWW-Authenticate` should be provided
	tokenCookie, err := request.Cookie(cookieName)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("request cookie: %w", err), cookieName)
		if errors.Is(err, http.ErrNoCookie) {
			return nil, &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnauthorized,
					"Missing token cookie.",
					nil,
				),
			}
		}
		return nil, &muxResponseError.ResponseError{ServerError: wrappedErr}
	}
	if tokenCookie == nil {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilCookie),
		}
	}

	tokenString := tokenCookie.Value
	if tokenString == "" {
		return nil, &muxResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusUnauthorized,
				"Empty token cookie.",
				nil,
			),
		}
	}

	signingKey := c.SigningKey
	if len(signingKey) == 0 {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrEmptySigningKey),
		}
	}

	claims, err := motmedelJwt.Validate(tokenString, c.SigningKey)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(
			fmt.Errorf("validate token: %w", err),
			tokenString,
			c.SigningKey,
		)
		if errors.Is(err, motmedelErrors.ErrValidationError) {
			return nil, &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnauthorized,
					"Invalid token cookie.",
					nil,
				),
			}
		}
		return nil, &muxResponseError.ResponseError{ServerError: wrappedErr}
	}

	return &muxUtilsJwt.TokenClaims{RegisteredClaims: claims, TokenString: tokenString}, nil
}
