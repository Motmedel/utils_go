package schema

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"

	"github.com/Motmedel/jsonschema"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxUtils "github.com/Motmedel/utils_go/pkg/http/mux/utils"
	muxUtilsJson "github.com/Motmedel/utils_go/pkg/http/mux/utils/json"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	motmedelJsonSchema "github.com/Motmedel/utils_go/pkg/json/schema"
	"github.com/Motmedel/utils_go/pkg/utils"
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

func (bodyParser *JsonSchemaBodyParser[T]) Parse(request *http.Request, body []byte) (T, *response_error.ResponseError) {
	var zero T

	schema := bodyParser.Schema
	if schema == nil {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(ErrNilSchema)}
	}

	dataMap, responseError := muxUtilsJson.ParseJsonBody[map[string]any](body)
	if responseError != nil {
		return zero, responseError
	}

	if err := Validate(dataMap, schema); err != nil {
		wrappedErr := motmedelErrors.New(fmt.Errorf("validate (input): %w", err), dataMap, schema)

		var validateError *ValidateError
		if errors.As(err, &validateError) {
			return zero, &response_error.ResponseError{
				// TODO: The error messages could be made nicer.
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnprocessableEntity,
					"Invalid body.",
					map[string][]string{"errors": slices.Collect(maps.Values(validateError.ListErrors))},
				),
				ClientError: wrappedErr,
			}
		}

		return zero, &response_error.ResponseError{ServerError: wrappedErr}
	}

	parser := bodyParser.BodyParser
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

func NewWithSchema[T any](schema *jsonschema.Schema) *JsonSchemaBodyParser[T] {
	return &JsonSchemaBodyParser[T]{BodyParser: muxUtils.MakeJsonBodyParser[T](), Schema: schema}
}

func New[T any]() (*JsonSchemaBodyParser[T], error) {
	var t T
	schema, err := motmedelJsonSchema.New[T](t)
	if err != nil {
		return nil, fmt.Errorf("schema new: %w", err)
	}

	return NewWithSchema[T](schema), nil
}
