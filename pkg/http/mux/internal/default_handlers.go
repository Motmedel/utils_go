package internal

import (
	"context"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
)

func DefaultResponseErrorHandler(
	ctx context.Context,
	responseError *muxTypesResponseError.ResponseError,
	responseWriter *muxTypesResponse.ResponseWriter,
) {
	if responseError == nil {
		return
	}

	// TODO: Make a function without the default argument...
	logger := motmedelLog.Logger{
		Logger: motmedelLog.GetLoggerFromCtxWithDefault(ctx, nil),
	}

	if responseWriter == nil {
		logger.Error(
			"The response writer is nil.",
			motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilResponseWriter),
		)
		return
	}

	switch responseErrorType := responseError.Type(); responseErrorType {
	case muxTypesResponseError.ResponseErrorType_ClientError:
		defer func() {
			logger.Warning("A client error occurred.", responseError.ClientError)
		}()
	case muxTypesResponseError.ResponseErrorType_ServerError:
		defer func() {
			logger.Error("A server error occurred.", responseError.ServerError)
		}()
	case muxTypesResponseError.ResponseErrorType_Invalid:
		logger.Error(
			"An invalid response error type was encountered.",
			motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrInvalidResponseError, responseError),
		)
		return
	default:
		logger.Error(
			"An unexpected response error type was encountered.",
			motmedelErrors.MakeErrorWithStackTrace(
				fmt.Errorf("%w: %v", muxErrors.ErrUnexpectedResponseErrorType, responseErrorType),
			),
		)
		return
	}

	if responseWriter.WriteHeaderCalled {
		return
	}

	response, err := responseError.MakeResponse()
	if err != nil {
		logger.Error(
			"An error occurred when making a response from a response error.",
			motmedelErrors.MakeError(
				fmt.Errorf("make response error response: %w", err),
				responseError,
			),
		)
		return
	}

	if err := responseWriter.WriteResponse(response); err != nil {
		logger.Error(
			"An error occurred when writing an error response.",
			motmedelErrors.MakeError(
				fmt.Errorf("write response: %w", err),
				responseError,
			),
		)
	}
}
