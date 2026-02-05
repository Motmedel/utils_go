package cors_configurator

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	"github.com/Motmedel/utils_go/pkg/interfaces/comparer"
	"github.com/Motmedel/utils_go/pkg/net/domain_breakdown"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type Configurator struct {
	AllowedOrigins         []string
	AllowedOriginsComparer comparer.Comparer[string]
	RegisteredDomain       string

	Headers       []string
	Credentials   bool
	MaxAge        int
	ExposeHeaders []string
}

func (c *Configurator) Parse(request *http.Request) (*motmedelHttpTypes.CorsConfiguration, *response_error.ResponseError) {
	if request == nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestHeader := request.Header
	if requestHeader == nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	origin := requestHeader.Get("Origin")
	if origin == "" {
		return nil, nil
	}

	var matchedAllowedOrigin string
	for _, allowedOrigin := range c.AllowedOrigins {
		if strings.EqualFold(origin, allowedOrigin) {
			matchedAllowedOrigin = allowedOrigin
			break
		}
	}

	if matchedAllowedOrigin != "" && !utils.IsNil(c.AllowedOriginsComparer) {
		if ok, _ := c.AllowedOriginsComparer.Compare(origin); ok {
			matchedAllowedOrigin = origin
		}
	}

	registeredDomain := c.RegisteredDomain
	if matchedAllowedOrigin == "" && registeredDomain != "" {
		parsedOrigin, err := url.Parse(origin)
		if err != nil {
			return nil, &response_error.ResponseError{
				ClientError: motmedelErrors.NewWithTrace(fmt.Errorf("url parse (origin): %w", err), origin),
				ProblemDetail: problem_detail.New(
					http.StatusBadRequest,
					problem_detail_config.WithDetail("Invalid Origin header."),
				),
			}
		}

		originHostname := parsedOrigin.Hostname()
		originDomainBreakdown := domain_breakdown.GetDomainBreakdown(originHostname)
		if originDomainBreakdown == nil {
			return nil, &response_error.ResponseError{
				ProblemDetail: problem_detail.New(
					http.StatusBadRequest,
					problem_detail_config.WithDetail("Invalid Origin header hostname."),
				),
			}
		}

		if strings.EqualFold(originDomainBreakdown.RegisteredDomain, registeredDomain) {
			matchedAllowedOrigin = origin
		}
	}

	if matchedAllowedOrigin == "" {
		return nil, nil
	}

	return &motmedelHttpTypes.CorsConfiguration{
		Origin:        matchedAllowedOrigin,
		Headers:       c.Headers,
		Credentials:   c.Credentials,
		MaxAge:        c.MaxAge,
		ExposeHeaders: c.ExposeHeaders,
	}, nil
}
