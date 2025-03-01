package schema

import (
	"errors"
	"fmt"
	"github.com/Motmedel/jsonschema"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser"
	bodyParserJson "github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser/json"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelInterfaces "github.com/Motmedel/utils_go/pkg/interfaces"
	motmedelJsonSchema "github.com/Motmedel/utils_go/pkg/json/schema"
	"maps"
	"net/http"
	"slices"
)

var (
	ErrNilSchema               = errors.New("nil schema")
	ErrNilDataMapPointer       = errors.New("nil data mad pointer")
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
	Schema    *jsonschema.Schema
	Processor body_parser.BodyProcessor[*T]
}

func (bodyParser *JsonSchemaBodyParser[T]) Parse(body []byte) (any, *response_error.ResponseError) {
	if bodyParser.Schema == nil {
		return nil, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(ErrNilSchema)}
	}

	dataMapPtr, responseError := bodyParserJson.ParseJsonBody[map[string]any](body)
	if responseError != nil {
		return nil, responseError
	}
	if dataMapPtr == nil {
		return nil, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(ErrNilDataMapPointer)}
	}

	if err := Validate(*dataMapPtr, bodyParser.Schema); err != nil {
		wrappedErr := motmedelErrors.New(fmt.Errorf("validate (input): %w", err), *dataMapPtr, bodyParser.Schema)

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
	result, responseError = bodyParser.BodyParser.Parse(body)
	if responseError != nil {
		return nil, responseError
	}

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

	if processor := bodyParser.Processor; processor != nil {
		result, responseError = processor.Process(result)
		if responseError != nil {
			return nil, responseError
		}
	}

	return result, nil
}

func NewWithSchema[T any](schema *jsonschema.Schema) body_parser.BodyParser[T] {
	return &JsonSchemaBodyParser[T]{BodyParser: bodyParserJson.New[T](), Schema: schema}
}

func New[T any](t T) (body_parser.BodyParser[T], error) {
	schema, err := motmedelJsonSchema.New[T](t)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("schema new: %w", err), t)
	}

	return NewWithSchema[T](schema), nil
}
