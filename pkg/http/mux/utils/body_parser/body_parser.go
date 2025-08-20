package body_parser

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/Motmedel/utils_go/pkg/interfaces/validatable"
	"github.com/Motmedel/utils_go/pkg/utils"
	"net/http"
)

func MakeValidatableBodyParser[T validatable.Validatable]() body_processor.BodyProcessor[T, T] {
	return body_processor.BodyProcessorFunction[T, T](
		func(v T) (T, *response_error.ResponseError) {
			var zero T

			if utils.IsNil(v) {
				return zero, nil
			}

			if err := v.Validate(); err != nil {
				wrappedErr := motmedelErrors.New(fmt.Errorf("validate: %w", err))
				if errors.Is(wrappedErr, motmedelErrors.ErrValidationError) {
					return zero, &response_error.ResponseError{
						ClientError: err,
						ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
							http.StatusBadRequest,
							"The input did not pass validation.",
							nil,
						),
					}
				}
			}

			return v, nil
		},
	)
}
