package json_body_parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
)

type Parser[T any] struct{}

func (p *Parser[T]) Parse(_ *http.Request, body []byte) (T, *response_error.ResponseError) {
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
		}

		return target, &response_error.ResponseError{ServerError: wrappedErr}
	}

	return target, nil
}

func New[T any]() *Parser[T] {
	return &Parser[T]{}
}
