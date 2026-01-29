package jwt_extractor

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/mismatch_error"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	authenticatorPkg "github.com/Motmedel/utils_go/pkg/interfaces/authenticator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/token/authenticated_token"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type Parser[T request_parser.RequestParser[string]] struct {
	TokenExtractor T
	Authenticators []authenticatorPkg.Authenticator[*authenticated_token.Token, string]
}

func (p *Parser[T]) Parse(request *http.Request) (*authenticated_token.Token, *muxResponseError.ResponseError) {
	if request == nil {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	tokenExtractor := p.TokenExtractor
	if utils.IsNil(tokenExtractor) {
		return nil, &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(
				fmt.Errorf("%w (token extractor)", muxErrors.ErrNilRequestParser),
			),
		}
	}

	tokenString, responseError := tokenExtractor.Parse(request)
	if responseError != nil {
		return nil, responseError
	}
	if tokenString == "" {
		return nil, &muxResponseError.ResponseError{
			ProblemDetail: problem_detail.New(
				http.StatusUnauthorized,
				problem_detail_config.WithDetail("Empty token."),
			),
		}
	}

	var authenticatedToken *authenticated_token.Token
	var waitGroup sync.WaitGroup

	authenticatorErrs := make([]error, len(p.Authenticators))
	for i, authenticator := range p.Authenticators {
		if utils.IsNil(authenticator) {
			continue
		}

		waitGroup.Go(
			func() {
				token, err := authenticator.Authenticate(request.Context(), tokenString)
				if err != nil {
					authenticatorErrs[i] = err
					return
				}

				authenticatedToken = token
				return
			},
		)
	}

	waitGroup.Wait()

	if authenticatedToken != nil {
		return authenticatedToken, nil
	}

	for _, err := range authenticatorErrs {
		if err == nil {
			continue
		}

		if e, ok := motmedelErrors.AsType[*mismatch_error.Error](err); ok && e.Field == "sub" {
			return nil, &muxResponseError.ResponseError{
				ClientError: err,
				ProblemDetail: problem_detail.New(
					http.StatusForbidden,
					problem_detail_config.WithDetail("The subject is not allowed to access this resource."),
				),
			}
		} else if motmedelErrors.IsAny(err, motmedelCryptoErrors.ErrSignatureMismatch, motmedelErrors.ErrValidationError) {
			return nil, &muxResponseError.ResponseError{
				ClientError: err,
				ProblemDetail: problem_detail.New(
					http.StatusUnauthorized,
					problem_detail_config.WithDetail("Invalid token."),
				),
			}
		}
	}

	return nil, &muxResponseError.ResponseError{ServerError: errors.Join(authenticatorErrs...)}
}

func New[T request_parser.RequestParser[string]](
	tokenExtractor T,
	authenticators ...authenticatorPkg.Authenticator[*authenticated_token.Token, string],
) (*Parser[T], error) {
	if utils.IsNil(tokenExtractor) {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w (token extractor)", muxErrors.ErrNilRequestParser))
	}

	return &Parser[T]{TokenExtractor: tokenExtractor, Authenticators: authenticators}, nil
}
