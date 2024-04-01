package utils_go

import (
	"fmt"
	utils_go "github.com/Motmedel/utils_go/utils_go/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

func RequestWrapper(
	responseWriter http.ResponseWriter,
	request *http.Request,
	f func(*[]zap.Field) (error, *utils_go.ProblemDetail),
	logger *zap.Logger,
) error {
	var wrapperFuncErr error
	var problemDetail *utils_go.ProblemDetail

	logFields := make([]zap.Field, 0, 32)

	clientIpAddress, clientPort, err := SplitAddress(request.RemoteAddr)
	if err != nil {
		problemDetail = utils_go.MakeInternalServerErrorProblemDetail()
	} else {
		logFields = append(
			logFields,
			zap.String("client.ip", clientIpAddress),
			zap.Int("client.port", clientPort),
		)

		if request.Header.Get("Content-Type") != "application/json" {
			problemDetail = &utils_go.ProblemDetail{
				Type:     "about:blank",
				Title:    "Unsupported Media Type",
				Status:   http.StatusUnsupportedMediaType,
				Detail:   "Content-Type is not application/json.",
				Instance: uuid.New().String(),
			}
		} else {
			wrapperFuncErr, problemDetail = f(&logFields)
			if wrapperFuncErr != nil && problemDetail == nil {
				problemDetail = utils_go.MakeInternalServerErrorProblemDetail()
			}
		}
	}

	if problemDetail != nil {
		if err != nil {
			logFields = append(
				logFields,
				zap.Error(err),
				zap.Stack("error.stack_trace"),
				zap.String("error.id", problemDetail.Instance),
			)
			logger.Error("A server error occurred in the wrapper prelude.", logFields...)
		} else if wrapperFuncErr != nil {
			logFields = append(
				logFields,
				zap.Error(wrapperFuncErr),
				zap.Stack("error.stack_trace"),
				zap.String("error.id", problemDetail.Instance),
			)
			logger.Error("A server error occurred in a wrapped function", logFields...)
		} else {
			logFields = append(
				logFields,
				zap.String("error.id", problemDetail.Instance),
				zap.String("error.message", problemDetail.Title+": "+problemDetail.Detail),
				zap.Any("problem_detail", problemDetail),
			)
			logger.Error("A client error occurred", logFields...)
		}

		responseWriter.Header().Set("Content-Type", "application/problem+json")
		responseWriter.Header().Set("X-Content-Type-Options", "nosniff")
		responseWriter.WriteHeader(problemDetail.Status)

		problemDetailString, err := problemDetail.String()
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(responseWriter, problemDetailString); err != nil {
			return err
		}
	}

	return nil
}
