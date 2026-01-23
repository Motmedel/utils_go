package request_parser

import (
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	processorPkg "github.com/Motmedel/utils_go/pkg/http/mux/types/processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
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

	processedResult, responseError := processor.Process(request.Context(), result)
	if responseError != nil {
		return zero, responseError
	}

	return processedResult, nil
}

func NewWithProcessor[T any, U any](requestParser RequestParser[T], processor processorPkg.Processor[U, T]) *RequestParserWithProcessor[T, U] {
	return &RequestParserWithProcessor[T, U]{
		RequestParser: requestParser,
		Processor:     processor,
	}
}
