package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser"
	processorPkg "github.com/Motmedel/utils_go/pkg/http/mux/interfaces/processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/request_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxUtilsJson "github.com/Motmedel/utils_go/pkg/http/mux/utils/json"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/Motmedel/utils_go/pkg/interfaces/urler"
	"github.com/Motmedel/utils_go/pkg/interfaces/validatable"
	motmedelUtils "github.com/Motmedel/utils_go/pkg/utils"
)

func getParsed[T any](ctx context.Context, key any) (T, error) {
	value, err := motmedelContext.GetContextValue[T](ctx, key)
	if err != nil {
		return value, fmt.Errorf("get context value: %w", err)
	}

	return value, nil
}

func getNonZeroParsed[T comparable](ctx context.Context, key any) (T, error) {
	value, err := motmedelContext.GetNonZeroContextValue[T](ctx, key)
	if err != nil {
		return value, fmt.Errorf("get non zero context value: %w", err)
	}

	return value, nil
}

func GetServerNonZeroContextValue[T comparable](ctx context.Context, key any) (T, *response_error.ResponseError) {
	value, err := getNonZeroParsed[T](ctx, key)
	if err != nil {
		var zero T
		return zero, &response_error.ResponseError{
			ServerError: fmt.Errorf("get non zero context value: %w", err),
		}
	}

	return value, nil
}

func GetServerContextValue[T any](ctx context.Context, key any) (T, *response_error.ResponseError) {
	value, err := getParsed[T](ctx, key)
	if err != nil {
		var zero T
		return zero, &response_error.ResponseError{
			ServerError: fmt.Errorf("get context value: %w", err),
		}
	}

	return value, nil
}

func GetParsedRequestBody[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetServerParsedRequestBody[T any](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerContextValue[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetNonZeroParsedRequestBody[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetServerNonZeroParsedRequestBody[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetParsedRequestHeaders[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetNonZeroParsedRequestHeaders[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetServerNonZeroParsedRequestHeaders[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetParsedRequestUrl[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func GetServerNonZeroParsedRequestUrl[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func GetNonZeroParsedRequestUrl[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func MakeValidatableProcessor[T validatable.Validatable]() processorPkg.Processor[T, T] {
	return processorPkg.ProcessorFunction[T, T](
		func(v T) (T, *response_error.ResponseError) {
			var zero T

			if motmedelUtils.IsNil(v) {
				return zero, nil
			}

			if err := v.Validate(); err != nil {
				wrappedErr := motmedelErrors.New(fmt.Errorf("validate: %w", err))
				if errors.Is(wrappedErr, motmedelErrors.ErrValidationError) {
					return zero, &response_error.ResponseError{
						ClientError: err,
						ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
							http.StatusBadRequest,
							"The body did not pass validation.",
							nil,
						),
					}
				}
			}

			return v, nil
		},
	)
}

func MakeJsonBodyParser[T any]() body_parser.BodyParser[T] {
	return body_parser.BodyParserFunction[T](
		func(request *http.Request, body []byte) (T, *response_error.ResponseError) {
			return muxUtilsJson.ParseJsonBody[T](body)
		},
	)
}

type RequestParserWithProcessor[T any, U any] struct {
	RequestParser request_parser.RequestParser[T]
	Processor     processorPkg.Processor[U, T]
}

func (p *RequestParserWithProcessor[T, U]) Parse(request *http.Request) (U, *response_error.ResponseError) {
	var zero U

	requestParser := p.RequestParser
	if motmedelUtils.IsNil(requestParser) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilRequestParser)}
	}

	processor := p.Processor
	if motmedelUtils.IsNil(processor) {
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

type RequestParserWithUrlProcessor[T urler.StringURLer, U *url.URL] struct {
	request_parser.RequestParser[T]
}

func (p *RequestParserWithUrlProcessor[T, U]) Parse(request *http.Request) (U, *response_error.ResponseError) {
	var zero U

	requestParser := p.RequestParser
	if motmedelUtils.IsNil(requestParser) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilRequestParser)}
	}

	result, responseError := requestParser.Parse(request)
	if responseError != nil {
		return zero, responseError
	}
	if motmedelUtils.IsNil(result) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(urler.ErrNilStringUrler)}
	}

	urlString := result.URL()
	if urlString == "" {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				"Empty url.",
				nil,
			),
		}
	}

	parsedUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				"Malformed url.",
				nil,
			),
			ClientError: motmedelErrors.New(err, urlString),
		}
	}

	return parsedUrl, nil
}

type BodyParserWithProcessor[T any, U any] struct {
	BodyParser body_parser.BodyParser[T]
	Processor  processorPkg.Processor[U, T]
}

func (p *BodyParserWithProcessor[T, U]) Parse(request *http.Request, body []byte) (U, *response_error.ResponseError) {
	var zero U

	bodyParser := p.BodyParser
	if motmedelUtils.IsNil(bodyParser) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilBodyParser)}
	}

	processor := p.Processor
	if motmedelUtils.IsNil(processor) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilProcessor)}
	}

	result, responseError := bodyParser.Parse(request, body)
	if responseError != nil {
		return zero, responseError
	}

	processedResult, responseError := processor.Process(result)
	if responseError != nil {
		return zero, responseError
	}

	return processedResult, nil
}
