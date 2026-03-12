package json_schema_body_parser

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser/json_body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	"github.com/Motmedel/utils_go/pkg/utils"
	jsonschemaErrors "github.com/altshiftab/jsonschema/pkg/errors"
	"github.com/altshiftab/jsonschema/pkg/jsonschema"
)

var jsonMapBodyParser = json_body_parser.New[map[string]any]()
var jsonArrayBodyParser = json_body_parser.New[[]any]()

type Parser[T any] struct {
	schema     *jsonschema.Schema
	bodyParser body_parser.BodyParser[T]
}

func (p *Parser[T]) Parse(request *http.Request, body []byte) (T, *response_error.ResponseError) {
	var zero T

	schema := p.schema
	if schema == nil {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(jsonschemaErrors.ErrNilSchema)}
	}

	bodyParser := p.bodyParser
	if utils.IsNil(bodyParser) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilBodyParser)}
	}

	var data any
	var responseError *response_error.ResponseError

	if reflect.TypeFor[T]().Kind() == reflect.Slice {
		data, responseError = jsonArrayBodyParser.Parse(request, body)
	} else {
		data, responseError = jsonMapBodyParser.Parse(request, body)
	}
	if responseError != nil {
		return zero, responseError
	}

	if err := schema.Validate(data); err != nil {
		wrappedErr := motmedelErrors.New(fmt.Errorf("validate (input): %w", err), data, schema)

		var validateError *jsonschemaErrors.ValidateError
		if errors.As(err, &validateError) {
			return zero, &response_error.ResponseError{
				ProblemDetail: problem_detail.New(
					http.StatusUnprocessableEntity,
					problem_detail_config.WithDetail("Invalid body."),
					problem_detail_config.WithExtension(map[string]any{"errors": validateError.Errors}),
				),
				ClientError: wrappedErr,
			}
		}

		return zero, &response_error.ResponseError{ServerError: wrappedErr}
	}

	var result T
	result, parseResponseError := bodyParser.Parse(request, body)
	if parseResponseError != nil {
		return zero, parseResponseError
	}

	return result, nil
}

func NewWithSchema[T any](schema *jsonschema.Schema) (*Parser[T], error) {
	if schema == nil {
		return nil, motmedelErrors.NewWithTrace(jsonschemaErrors.ErrNilSchema)
	}

	return &Parser[T]{schema: schema, bodyParser: json_body_parser.New[T]()}, nil
}

func New[T any]() (*Parser[T], error) {
	schema, err := jsonschema.NewFromType[T]()
	if err != nil {
		return nil, fmt.Errorf("schema new: %w", err)
	}

	return NewWithSchema[T](schema)
}
