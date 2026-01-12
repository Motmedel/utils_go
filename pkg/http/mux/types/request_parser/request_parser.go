package request_parser

import (
	"fmt"
	"net/http"
	"net/url"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	processorPkg "github.com/Motmedel/utils_go/pkg/http/mux/types/processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/url_processor_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/Motmedel/utils_go/pkg/interfaces/urler"
	"github.com/Motmedel/utils_go/pkg/net/domain_breakdown"
	"github.com/Motmedel/utils_go/pkg/net/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type RequestParser[T any] interface {
	Parse(*http.Request) (T, *muxTypesResponseError.ResponseError)
}

type RequestParserFunction[T any] func(r *http.Request) (T, *muxTypesResponseError.ResponseError)

func (f RequestParserFunction[T]) Parse(request *http.Request) (T, *muxTypesResponseError.ResponseError) {
	return f(request)
}

func New[T any](f func(r *http.Request) (T, *muxTypesResponseError.ResponseError)) RequestParser[T] {
	return RequestParserFunction[T](f)
}

type RequestParserWithProcessor[T any, U any] struct {
	RequestParser RequestParser[T]
	Processor     processorPkg.Processor[U, T]
}

func (p *RequestParserWithProcessor[T, U]) Parse(request *http.Request) (U, *response_error.ResponseError) {
	var zero U

	requestParser := p.RequestParser
	if utils.IsNil(requestParser) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilRequestParser)}
	}

	processor := p.Processor
	if utils.IsNil(processor) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilProcessor)}
	}

	result, responseError := requestParser.Parse(request)
	if responseError != nil {
		return zero, responseError
	}

	processedResult, responseError := processor.Process(result)
	if responseError != nil {
		return zero, responseError
	}

	return processedResult, nil
}

func WithProcessor[T any, U any](requestParser RequestParser[T], processor processorPkg.Processor[U, T]) *RequestParserWithProcessor[T, U] {
	return &RequestParserWithProcessor[T, U]{
		RequestParser: requestParser,
		Processor:     processor,
	}
}

type RequestParserWithUrlProcessor[T urler.StringURLer] struct {
	RequestParser[T]
	Config *url_processor_config.Config
}

func (p *RequestParserWithUrlProcessor[T]) Parse(request *http.Request) (*url.URL, *response_error.ResponseError) {
	requestParser := p.RequestParser
	if utils.IsNil(requestParser) {
		return nil, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilRequestParser)}
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
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilUrlProcessorConfig),
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
				ClientError: motmedelErrors.NewWithTrace(errors.ErrNilDomainBreakdown),
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

func WithUrlProcessor[T urler.StringURLer](requestParser RequestParser[T], options ...url_processor_config.Option) *RequestParserWithUrlProcessor[T] {
	return &RequestParserWithUrlProcessor[T]{
		RequestParser: requestParser,
		Config:        url_processor_config.New(options...),
	}
}
