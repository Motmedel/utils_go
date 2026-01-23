package cookie_extractor

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/cookie_extractor/cookie_extractor_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
)

var (
	ErrEmptyName = errors.New("empty name")
)

type Parser struct {
	Name   string
	config *cookie_extractor_config.Config
}

func (p *Parser) Parse(request *http.Request) (string, *response_error.ResponseError) {
	if request == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	name := p.Name
	if name == "" {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(ErrEmptyName),
		}
	}

	cookie, err := request.Cookie(name)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("request cookie: %w", err), name)
		if errors.Is(err, http.ErrNoCookie) {
			return "", &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					p.config.ProblemDetailStatusCode,
					p.config.ProblemDetailText,
					map[string]string{"cookie": name},
				),
			}
		}
		return "", &muxResponseError.ResponseError{ServerError: wrappedErr}
	}
	if cookie == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilCookie),
		}
	}

	return cookie.Value, nil
}

func New(name string, options ...cookie_extractor_config.Option) (*Parser, error) {
	if name == "" {
		return nil, motmedelErrors.NewWithTrace(ErrEmptyName)
	}

	return &Parser{Name: name, config: cookie_extractor_config.New(options...)}, nil
}

func NewTokenCookieExtractor(name string, options ...cookie_extractor_config.Option) (*Parser, error) {
	return New(
		name,
		slices.Concat(
			[]cookie_extractor_config.Option{
				cookie_extractor_config.WithProblemDetailStatusCode(http.StatusUnauthorized),
				cookie_extractor_config.WithProblemDetailText("Missing token cookie."),
			},
			options,
		)...,
	)
}
