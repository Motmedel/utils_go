package json

import (
	"encoding/json"
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"net/http"
)

func ParseJsonBody[T any](_ *http.Request, body []byte) (*T, *response_error.ResponseError) {
	var target T

	if err := json.Unmarshal(body, &target); err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal: %w", err), body)

		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			return nil, &response_error.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnprocessableEntity,
					"Invalid body. The value is not appropriate for the JSON type.",
					nil,
				),
			}
		} else {
			return nil, &response_error.ResponseError{ServerError: wrappedErr}
		}
	}

	return &target, nil
}

func New[T any]() body_parser.BodyParser[T] {
	return body_parser.BodyParserFunction[T](ParseJsonBody[T])
}
