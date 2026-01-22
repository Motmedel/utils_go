package url_allower

import (
	"fmt"
	"net/http"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelMuxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/url_allower/url_allower_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/Motmedel/utils_go/pkg/interfaces/urler"
	"github.com/Motmedel/utils_go/pkg/net/domain_breakdown"
	motmedelNetErrors "github.com/Motmedel/utils_go/pkg/net/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type UrlAllower[T urler.StringURLer] struct {
	request_parser.RequestParser[T]
	Config *url_allower_config.Config
}

func (p *UrlAllower[T]) Parse(request *http.Request) (*url.URL, *response_error.ResponseError) {
	requestParser := p.RequestParser
	if utils.IsNil(requestParser) {
		return nil, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(motmedelMuxErrors.ErrNilRequestParser)}
	}

	result, responseError := requestParser.Parse(request)
	if responseError != nil {
		return nil, responseError
	}
	if utils.IsNil(result) {
		return nil, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(urler.ErrNilStringUrler)}
	}

	urlString := result.URL()
	if urlString == "" {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail("Empty url.", nil),
		}
	}

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail("Malformed url.", nil),
			ClientError:   motmedelErrors.NewWithTrace(fmt.Errorf("url parse: %w", err), urlString),
		}
	}

	config := p.Config
	if config == nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelMuxErrors.ErrNilUrlProcessorConfig),
		}
	}

	parsedUrlHostname := parsedUrl.Hostname()
	if !(config.AllowLocalhost && parsedUrlHostname == "localhost") {
		domainBreakdown := domain_breakdown.GetDomainBreakdown(parsedUrlHostname)
		if domainBreakdown == nil {
			return nil, &response_error.ResponseError{
				ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
					"Malformed url hostname; not a domain.",
					nil,
				),
				ClientError: motmedelErrors.NewWithTrace(motmedelNetErrors.ErrNilDomainBreakdown),
			}
		}

		if len(config.AllowedDomains) > 0 || len(config.AllowedRegisteredDomains) > 0 {
			var allowed bool

			registeredDomain := domainBreakdown.RegisteredDomain
			for _, domain := range config.AllowedRegisteredDomains {
				if registeredDomain == domain {
					allowed = true
					break
				}
			}

			if !allowed {
				for _, domain := range config.AllowedDomains {
					if domain == parsedUrlHostname {
						allowed = true
						break
					}
				}
			}

			if !allowed {
				return nil, &response_error.ResponseError{
					ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
						"The url hostname does not match any allowed domain.",
						nil,
					),
				}
			}
		}
	}

	return parsedUrl, nil
}

func New[T urler.StringURLer](requestParser request_parser.RequestParser[T], options ...url_allower_config.Option) *UrlAllower[T] {
	return &UrlAllower[T]{
		RequestParser: requestParser,
		Config:        url_allower_config.New(options...),
	}
}
