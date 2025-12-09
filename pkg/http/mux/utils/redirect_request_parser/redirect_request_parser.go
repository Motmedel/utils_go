package redirect_request_parser

import (
	"errors"
	"net/http"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/request_parser"
	muxInternalVhostMux "github.com/Motmedel/utils_go/pkg/http/mux/internal/vhost_mux"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/redirect_request_parser/config"
	motmedelNetErrors "github.com/Motmedel/utils_go/pkg/net/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type RequestParser[T any] struct {
	request_parser.RequestParser[T]
	RedirectUrl       *url.URL
	RedirectParameter string
	RequireProto      bool
}

func (parser *RequestParser[T]) Parse(request *http.Request) (T, *response_error.ResponseError) {
	var zero T

	requestParser := parser.RequestParser
	if utils.IsNil(requestParser) {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilRequestParser),
		}
	}

	if parser.RedirectUrl == nil {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelNetErrors.ErrNilUrl),
		}
	}

	requestHeader := request.Header
	if requestHeader == nil {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	host := request.Host
	if host == "" {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrEmptyHost),
		}
	}

	requestUrl := request.URL
	if requestUrl == nil {
		return zero, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestUrl),
		}
	}

	// TODO: Try `Forwarded` first, then `X-Forwarded-Proto`.
	scheme := requestHeader.Get("X-Forwarded-Proto")
	if scheme == "" {
		if parser.RequireProto {
			return zero, &response_error.ResponseError{
				ServerError: motmedelErrors.NewWithTrace(errors.New("missing X-Forwarded-Proto header")),
			}
		}

		if request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	out, responseError := requestParser.Parse(request)
	if responseError == nil {
		return out, nil
	}

	if requestHeader.Get("Sec-Fetch-Mode") != "navigate" {
		return zero, responseError
	}

	problemDetail := responseError.ProblemDetail
	if problemDetail == nil {
		return zero, responseError
	}

	if problemDetail.Status != http.StatusUnauthorized {
		return zero, responseError
	}

	redirectParameter := parser.RedirectParameter
	if redirectParameter == "" {
		redirectParameter = config.DefaultParameterName
	}

	currentUrl := *request.URL
	currentUrl.Host = host
	currentUrl.Scheme = scheme

	redirectUrl := *parser.RedirectUrl
	query := redirectUrl.Query()
	query.Set(redirectParameter, currentUrl.String())
	redirectUrl.RawQuery = query.Encode()

	responseError.Headers = append(
		responseError.Headers,
		&response.HeaderEntry{
			Name:  "Location",
			Value: muxInternalVhostMux.HexEscapeNonASCII(redirectUrl.String()),
		},
	)

	return zero, responseError
}

func New[T any](
	requestParser request_parser.RequestParser[T],
	redirectUrl *url.URL,
	options ...config.Option,
) (*RequestParser[T], error) {
	if utils.IsNil(requestParser) {
		return nil, motmedelErrors.NewWithTrace(muxErrors.ErrNilRequestParser)
	}

	if redirectUrl == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelNetErrors.ErrNilUrl)
	}

	requestParserConfig := config.New(options...)

	return &RequestParser[T]{
		RequestParser:     requestParser,
		RedirectUrl:       redirectUrl,
		RedirectParameter: requestParserConfig.ParameterName,
		RequireProto:      requestParserConfig.RequireProto,
	}, nil
}
