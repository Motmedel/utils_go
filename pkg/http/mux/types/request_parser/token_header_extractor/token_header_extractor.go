package token_cookie_extractor

import (
	"fmt"
	"net/http"

	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/cookie_extractor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/cookie_extractor/cookie_extractor_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/token_cookie_extractor/token_cookie_extractor_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
)

type TokenCookieExtractor struct {
	Name   string
	config *token_cookie_extractor_config.Config
}

func (tce *TokenCookieExtractor) Parse(request *http.Request) (string, *response_error.ResponseError) {
	cookieExtractor, err := cookie_extractor.New(
		tce.config.CookieName,
		cookie_extractor_config.WithProblemDetailText(tce.config.ProblemDetailText),
		cookie_extractor_config.WithProblemDetailStatusCode(tce.config.ProblemDetailStatusCode),
	)
	if err != nil {
		return "", &response_error.ResponseError{ServerError: fmt.Errorf("cookie extractor new: %w", err)}
	}

	cookie, responseError := cookieExtractor.Parse(request)
	if responseError != nil {
		return "", responseError
	}

}

func New(name string, options ...token_cookie_extractor_config.Option) *TokenCookieExtractor {
	return &TokenCookieExtractor{config: token_cookie_extractor_config.New(options...)}
}
