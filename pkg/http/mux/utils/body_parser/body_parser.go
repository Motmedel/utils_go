package body_parser

import (
	"encoding/json"
	"errors"
	"github.com/Motmedel/jsonschema"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxTypes "github.com/Motmedel/utils_go/pkg/http/mux/types"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"maps"
	"net/http"
	"slices"
)

var (
	ErrNilSchema               = errors.New("nil schema")
	ErrNilEvaluationResult     = errors.New("nil evaluation result")
	ErrNilEvaluationResultList = errors.New("nil evaluation result list")
)

func MakeMuxJsonTypeValidator[T any](
	schema *jsonschema.Schema,
	extraCheck func(result T) *muxTypes.HandlerErrorResponse,
) (func(*http.Request, []byte) (any, *muxTypes.HandlerErrorResponse), error) {
	if schema == nil {
		return nil, ErrNilSchema
	}

	return func(_ *http.Request, body []byte) (any, *muxTypes.HandlerErrorResponse) {
		var dataMap map[string]any
		if err := json.Unmarshal(body, &dataMap); err != nil {
			var unmarshalTypeError *json.UnmarshalTypeError
			if errors.As(err, &unmarshalTypeError) {
				return nil, &muxTypes.HandlerErrorResponse{
					ClientError: err,
					ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
						http.StatusUnprocessableEntity,
						"Invalid body",
						map[string][]string{"errors": {"The value is not appropriate for the JSON type"}},
					),
				}
			} else {
				return nil, &muxTypes.HandlerErrorResponse{
					ServerError: &motmedelErrors.InputError{
						Message: "An error occurred when parsing data as a map as a step to validate the data.",
						Cause:   err,
						Input:   body,
					},
				}
			}
		}

		evaluationResult := schema.Validate(dataMap)
		if evaluationResult == nil {
			return nil, &muxTypes.HandlerErrorResponse{ServerError: ErrNilEvaluationResult}
		}

		evaluationResultList := evaluationResult.ToList()
		if evaluationResultList == nil {
			return nil, &muxTypes.HandlerErrorResponse{ServerError: ErrNilEvaluationResultList}
		}

		if errorsMap := evaluationResultList.Errors; len(errorsMap) != 0 {
			return nil, &muxTypes.HandlerErrorResponse{
				// TODO: The error messages could be made nicer.
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnprocessableEntity,
					"Invalid body",
					map[string][]string{"errors": slices.Collect(maps.Values(errorsMap))},
				),
			}
		} else {
			var result T

			if err := json.Unmarshal(body, &result); err != nil {
				return nil, &muxTypes.HandlerErrorResponse{
					ServerError: &motmedelErrors.InputError{
						Message: "An error occurred when parsing data as the resulting type.",
						Cause:   err,
						Input:   body,
					},
				}
			}

			if extraCheck != nil {
				if handlerErrorResponse := extraCheck(result); handlerErrorResponse != nil {
					return nil, handlerErrorResponse
				}
			}

			return result, nil
		}
	}, nil
}
