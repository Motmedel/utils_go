package json_schema_body_parser

import (
	"errors"
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxUtils "github.com/Motmedel/utils_go/pkg/http/mux/utils"
	muxUtilsJson "github.com/Motmedel/utils_go/pkg/http/mux/utils/json"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelJsonSchema "github.com/Motmedel/utils_go/pkg/json/schema"
	"github.com/Motmedel/utils_go/pkg/utils"
	jsonschemaErrors "github.com/altshiftab/jsonschema/pkg/errors"
	"github.com/altshiftab/jsonschema/pkg/jsonschema"
)

type Parser[T any] struct {
	body_parser.BodyParser[T]
	Schema *jsonschema.Schema
}

func (p *Parser[T]) Parse(request *http.Request, body []byte) (T, *response_error.ResponseError) {
	var zero T

	schema := p.Schema
	if schema == nil {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(jsonschemaErrors.ErrNilSchema)}
	}

	dataMap, responseError := muxUtilsJson.ParseJsonBody[map[string]any](body)
	if responseError != nil {
		return zero, responseError
	}

	if err := motmedelJsonSchema.Validate(dataMap, schema); err != nil {
		wrappedErr := motmedelErrors.New(fmt.Errorf("validate (input): %w", err), dataMap, schema)

		var validateError *motmedelJsonSchema.ValidateError
		if errors.As(err, &validateError) {
			return zero, &response_error.ResponseError{
				// TODO: The error messages could be made nicer.
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnprocessableEntity,
					"Invalid body.",
					map[string]any{"errors": validateError.Errors},
				),
				ClientError: wrappedErr,
			}
		}

		return zero, &response_error.ResponseError{ServerError: wrappedErr}
	}

	parser := p.BodyParser
	if utils.IsNil(parser) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilBodyParser)}
	}

	var result T
	result, responseError = parser.Parse(request, body)
	if responseError != nil {
		return zero, responseError
	}

	return result, nil
}

func NewWithSchema[T any](schema *jsonschema.Schema) *Parser[T] {
	return &Parser[T]{BodyParser: muxUtils.MakeJsonBodyParser[T](), Schema: schema}
}

func New[T any]() (*Parser[T], error) {
	schema, err := jsonschema.FromType[T]()
	if err != nil {
		return nil, fmt.Errorf("schema new: %w", err)
	}

	return NewWithSchema[T](schema), nil
}
