package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
)

func ParseJsonBody[T any](body []byte) (T, *response_error.ResponseError) {
	var target T

	if err := json.Unmarshal(body, &target); err != nil {
		wrappedErr := motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal: %w", err), body)

		var unmarshalTypeError *json.UnmarshalTypeError
		if errors.As(err, &unmarshalTypeError) {
			return target, &response_error.ResponseError{
				ClientError: wrappedErr,
				ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
					http.StatusUnprocessableEntity,
					"Invalid body. The value is not appropriate for the JSON type.",
					nil,
				),
			}
		} else {
			return target, &response_error.ResponseError{ServerError: wrappedErr}
		}
	}

	return target, nil
}

func New[T any]() body_parser.BodyParser[T] {
	return body_parser.BodyParserFunction[T](
		func(request *http.Request, body []byte) (T, *response_error.ResponseError) {
			return ParseJsonBody[T](body)
		},
	)
}
