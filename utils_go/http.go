package utils_go

import (
	"fmt"
	utils_go "github.com/Motmedel/utils_go/utils_go/types"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

func logClientError(logger *slog.Logger, problemDetail *utils_go.ProblemDetail, err error) {
	slogErrorAttributes := []any{
		slog.String("id", problemDetail.Instance),
	}

	if err != nil {
		slogErrorAttributes = append(slogErrorAttributes, slog.String("message", err.Error()))
	}

	logger.Error(
		"A client error occurred.",
		"problem_detail", problemDetail,
		slog.Group(
			"error",
			slogErrorAttributes...
		),
	)
}

func RequestWrapper(
	responseWriter http.ResponseWriter,
	request *http.Request,
	f func(*slog.Logger) (error, *utils_go.ProblemDetail),
	logger *slog.Logger,
) error {
	var wrapperFuncErr error
	var problemDetail *utils_go.ProblemDetail

	clientIpAddress, clientPort, err := SplitAddress(request.RemoteAddr)

	if err != nil {
		problemDetail = utils_go.MakeInternalServerErrorProblemDetail()
		logger.Error(
			"A server error occurred in the wrapper prelude.",
			slog.Group(
				"error",
				slog.String("message", err.Error()),
				slog.String("id", problemDetail.Instance),
			),
		)
	} else {
		logger = logger.With(
			slog.Group(
				"client",
				slog.String("ip", clientIpAddress),
				slog.Int("port", clientPort),
			),
		)

		if request.Header.Get("Content-Type") != "application/json" {
			problemDetail = &utils_go.ProblemDetail{
				Type:     "about:blank",
				Title:    "Unsupported Media Type",
				Status:   http.StatusUnsupportedMediaType,
				Detail:   "Content-Type is not application/json.",
				Instance: uuid.New().String(),
			}
			logClientError(logger, problemDetail, nil)
		} else {
			wrapperFuncErr, problemDetail = f(logger)
			if wrapperFuncErr != nil {
				if problemDetail == nil {
					problemDetail = utils_go.MakeInternalServerErrorProblemDetail()
					logger.Error(
						"A server error occurred in a wrapped function.",
						slog.Group(
							"error",
							slog.String("message", wrapperFuncErr.Error()),
							slog.String("id", problemDetail.Instance),
						),
					)
				} else {
					logClientError(logger, problemDetail, wrapperFuncErr)
				}
			}
		}
	}

	if problemDetail != nil {
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
