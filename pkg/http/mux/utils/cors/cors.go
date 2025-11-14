package cors

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"github.com/Motmedel/utils_go/pkg/net/domain_breakdown"
)

type Configurator struct {
	AllowedOrigins   []string
	RegisteredDomain string

	Headers       []string
	Credentials   bool
	MaxAge        int
	ExposeHeaders []string
}

func (configurator *Configurator) Parse(request *http.Request) (*motmedelHttpTypes.CorsConfiguration, *response_error.ResponseError) {
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
	for _, allowedOrigin := range configurator.AllowedOrigins {
		if strings.EqualFold(origin, allowedOrigin) {
			matchedAllowedOrigin = allowedOrigin
			break
		}
	}

	registeredDomain := configurator.RegisteredDomain
	if matchedAllowedOrigin == "" && registeredDomain != "" {
		parsedOrigin, err := url.Parse(origin)
		if err != nil {
			return nil, &response_error.ResponseError{
				ClientError: motmedelErrors.NewWithTrace(fmt.Errorf("url parse (origin): %w", err), origin),
				ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
					"Invalid Origin header.",
					nil,
				),
			}
		}

		originHostname := parsedOrigin.Hostname()
		originDomainBreakdown := domain_breakdown.GetDomainBreakdown(originHostname)
		if originDomainBreakdown == nil {
			return nil, &response_error.ResponseError{
				ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
					"Invalid Origin header hostname.",
					nil,
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
		Headers:       configurator.Headers,
		Credentials:   configurator.Credentials,
		MaxAge:        configurator.MaxAge,
		ExposeHeaders: configurator.ExposeHeaders,
	}, nil
}
