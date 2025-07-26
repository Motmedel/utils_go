package schema

import (
	"errors"
	"fmt"
	"github.com/Motmedel/jsonschema"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	bodyParserJson "github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser/json"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelInterfaces "github.com/Motmedel/utils_go/pkg/interfaces"
	motmedelJsonSchema "github.com/Motmedel/utils_go/pkg/json/schema"
	"github.com/Motmedel/utils_go/pkg/utils"
	"maps"
	"net/http"
	"slices"
)

var (
	ErrNilSchema               = errors.New("nil schema")
	ErrNilEvaluationResult     = errors.New("nil evaluation result")
	ErrNilEvaluationResultList = errors.New("nil evaluation result list")
)

type ValidateError struct {
	error
	ListErrors map[string]string
}

func Validate(dataMap map[string]any, schema *jsonschema.Schema) error {
	if schema == nil {
		return motmedelErrors.NewWithTrace(ErrNilSchema)
	}

	evaluationResult := schema.Validate(dataMap)
	if evaluationResult == nil {
		return motmedelErrors.NewWithTrace(ErrNilEvaluationResult)
	}

	evaluationResultList := evaluationResult.ToList()
	if evaluationResultList == nil {
		return motmedelErrors.NewWithTrace(ErrNilEvaluationResultList)
	}

	errorsMap := evaluationResultList.Errors
	if len(errorsMap) != 0 {
		return &ValidateError{error: motmedelErrors.ErrValidationError, ListErrors: errorsMap}
	}

	return nil
}

type JsonSchemaBodyParser[T any] struct {
	body_parser.BodyParser[T]
	Schema *jsonschema.Schema
}

func (bodyParser *JsonSchemaBodyParser[T]) Parse(request *http.Request, body []byte) (any, *response_error.ResponseError) {
	schema := bodyParser.Schema
	if schema == nil {
		return nil, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(ErrNilSchema)}
	}

	dataMap, responseError := bodyParserJson.ParseJsonBody[map[string]any](body)
	if responseError != nil {
		return nil, responseError
	}

	if err := Validate(dataMap, schema); err != nil {
		wrappedErr := motmedelErrors.New(fmt.Errorf("validate (input): %w", err), dataMap, schema)

		var validateError *ValidateError
		if errors.As(err, &validateError) {
			return nil, &response_error.ResponseError{
				// TODO: The error messages could be made nicer.
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnprocessableEntity,
					"Invalid body.",
					map[string][]string{"errors": slices.Collect(maps.Values(validateError.ListErrors))},
				),
				ClientError: wrappedErr,
			}
		}

		return nil, &response_error.ResponseError{ServerError: wrappedErr}
	}

	var result any
	result, responseError = bodyParser.BodyParser.Parse(request, body)
	if responseError != nil {
		return nil, responseError
	}

	// TODO: Not sure this is a nice solution.

	if validator, ok := result.(motmedelInterfaces.Validator); ok {
		if err := validator.Validate(); err != nil {
			wrappedErr := fmt.Errorf("validate (result): %w", err)

			if errors.Is(err, motmedelErrors.ErrValidationError) {
				return nil, &response_error.ResponseError{
					ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
						http.StatusUnprocessableEntity,
						"Invalid body.",
						map[string]string{"error": err.Error()},
					),
					ClientError: wrappedErr,
				}
			}

			return nil, &response_error.ResponseError{ServerError: wrappedErr}
		}
	}

	return result, nil
}

type JsonSchemaBodyParserWithProcessor[T any, U any] struct {
	JsonSchemaBodyParser[T]
	Processor body_processor.BodyProcessor[U, T]
}

func (bodyParser *JsonSchemaBodyParserWithProcessor[T, U]) Parse(request *http.Request, body []byte) (U, *response_error.ResponseError) {
	var zero U

	result, responseError := bodyParser.JsonSchemaBodyParser.Parse(request, body)
	if responseError != nil {
		return zero, responseError
	}

	if processor := bodyParser.Processor; !utils.IsNil(processor) {
		result, responseError = processor.Process(result)
		if responseError != nil {
			return zero, responseError
		}
	}

	return result, nil
}

func NewWithSchema[T any](schema *jsonschema.Schema) *JsonSchemaBodyParser[T] {
	return &JsonSchemaBodyParser[T]{
		BodyParser: bodyParserJson.New[T](),
		Schema:     schema,
	}
}

func New[T any]() (*JsonSchemaBodyParser[T], error) {
	var t T
	schema, err := motmedelJsonSchema.New[T](t)
	if err != nil {
		return nil, fmt.Errorf("schema new: %w", err)
	}

	return NewWithSchema[T](schema), nil
}

func NewWithProcessor[T any, U any](processor body_processor.BodyProcessor[U, T]) (*JsonSchemaBodyParserWithProcessor[T, U], error) {
	jsonSchemaBodyParser, err := New[T]()
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	return &JsonSchemaBodyParserWithProcessor[T, U]{
		JsonSchemaBodyParser: *jsonSchemaBodyParser,
		Processor:            processor,
	}, nil
}
