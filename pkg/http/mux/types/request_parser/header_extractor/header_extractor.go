package header_extractor

import (
	"errors"
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/header_extractor/header_extractor_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
)

var (
	ErrEmptyName = errors.New("empty name")
)

type Parser struct {
	Name   string
	config *header_extractor_config.Config
}

func (p *Parser) Parse(request *http.Request) (string, *response_error.ResponseError) {
	if request == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestHeader := request.Header
	if requestHeader == nil {
		return "", &muxResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	name := p.Name
	if name == "" {
		return "", &muxResponseError.ResponseError{ServerError: motmedelErrors.NewWithTrace(ErrEmptyName)}
	}

	headerValue, err := motmedelHttpUtils.GetSingleHeader(name, requestHeader)
	if err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("get single header: %w", err), name)

		if errors.Is(err, motmedelHttpErrors.ErrMissingHeader) {
			return "", &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.New(
					p.config.ProblemDetailStatusCode,
					problem_detail_config.WithDetail(p.config.ProblemDetailMissingText),
					problem_detail_config.WithExtension(map[string]any{"header": name}),
				),
			}
		} else if errors.Is(err, motmedelHttpErrors.ErrMultipleHeaderValues) {
			return "", &muxResponseError.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.New(
					p.config.ProblemDetailStatusCode,
					problem_detail_config.WithDetail(p.config.ProblemDetailMultipleText),
					problem_detail_config.WithExtension(map[string]any{"header": name}),
				),
			}
		}
		return "", &muxResponseError.ResponseError{ServerError: wrappedErr}
	}

	return headerValue, nil
}

func New(name string, options ...header_extractor_config.Option) (*Parser, error) {
	if name == "" {
		return nil, motmedelErrors.NewWithTrace(ErrEmptyName)
	}

	return &Parser{Name: name, config: header_extractor_config.New(options...)}, nil
}
