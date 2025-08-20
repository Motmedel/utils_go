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

var ValidatableProcessor = body_processor.BodyProcessorFunction[validatable.Validatable, validatable.Validatable](
	func(v validatable.Validatable) (validatable.Validatable, *response_error.ResponseError) {
		if utils.IsNil(v) {
			return nil, nil
		}

		if err := v.Validate(); err != nil {
			wrappedErr := motmedelErrors.New(fmt.Errorf("validate: %w", err))
			if errors.Is(wrappedErr, motmedelErrors.ErrValidationError) {
				return nil, &response_error.ResponseError{
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
