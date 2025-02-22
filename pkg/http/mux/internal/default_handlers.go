package internal

import (
	"context"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	muxContext "github.com/Motmedel/utils_go/pkg/http/mux/context"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"log/slog"
	"net/http"
)

func DefaultResponseErrorHandler(
	ctx context.Context,
	responseError *muxTypesResponseError.ResponseError,
	responseWriter *muxTypesResponse.ResponseWriter,
) {
	if responseError == nil {
		return
	}

	if responseWriter == nil {
		slog.ErrorContext(
			context.WithValue(
				ctx,
				motmedelErrors.ErrorContextKey,
				motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilResponseWriter),
			),
			"The response writer is nil.",
		)
		return
	}

	var errorId string

	switch responseErrorType := responseError.Type(); responseErrorType {
	case muxTypesResponseError.ResponseErrorType_ClientError:
		defer func() {
			clientError := motmedelErrors.MakeError(responseError.ClientError)
			clientError.Id = errorId
			slog.WarnContext(
				context.WithValue(ctx, motmedelErrors.ErrorContextKey, clientError),
				"A client error occurred.",
			)
		}()
	case muxTypesResponseError.ResponseErrorType_ServerError:
		defer func() {
			serverError := motmedelErrors.MakeError(responseError.ServerError)
			serverError.Id = errorId
			slog.ErrorContext(
				context.WithValue(ctx, motmedelErrors.ErrorContextKey, serverError),
				"A server error occurred.",
			)
		}()
	case muxTypesResponseError.ResponseErrorType_Invalid:
		slog.ErrorContext(
			context.WithValue(
				ctx,
				motmedelErrors.ErrorContextKey,
				motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrUnusableResponseError, responseError),
			),
			"An invalid response error type was encountered.",
		)
		return
	default:
		slog.ErrorContext(
			context.WithValue(
				ctx,
				motmedelErrors.ErrorContextKey,
				motmedelErrors.MakeErrorWithStackTrace(
					fmt.Errorf("%w: %v", muxErrors.ErrUnexpectedResponseErrorType, responseErrorType),
				),
			),
			"An unexpected response error type was encountered.",
		)
		return
	}

	if responseWriter.WriteHeaderCalled {
		return
	}

	problemDetail, err := responseError.GetEffectiveProblemDetail()
	if err != nil {
		slog.ErrorContext(
			context.WithValue(
				ctx,
				motmedelErrors.ErrorContextKey,
				motmedelErrors.MakeErrorWithStackTrace(
					fmt.Errorf("response error get effective problem detail: %w", err),
					responseError,
				),
			),
			"An error occurred when obtaining the effective response error problem detail.",
		)
		return
	}
	responseError.ProblemDetail = problemDetail
	errorId = problemDetail.Instance

	contentNegotiation, _ := ctx.Value(muxContext.ContentNegotiationContextKey).(*motmedelHttpTypes.ContentNegotiation)
	response, err := responseError.MakeResponse(contentNegotiation)
	if err != nil {
		slog.ErrorContext(
			context.WithValue(
				ctx,
				motmedelErrors.ErrorContextKey,
				motmedelErrors.MakeError(fmt.Errorf("make response error response: %w", err), responseError),
			),
			"An error occurred when making a response from a response error.",
		)
		return
	}

	var acceptEncoding *motmedelHttpTypes.AcceptEncoding
	if contentNegotiation != nil {
		acceptEncoding = contentNegotiation.AcceptEncoding
	}

	if err := responseWriter.WriteResponse(ctx, response, acceptEncoding); err != nil {
		slog.ErrorContext(
			context.WithValue(
				ctx,
				motmedelErrors.ErrorContextKey,
				motmedelErrors.MakeError(fmt.Errorf("write response: %w", err), responseError),
			),
			"An error occurred when writing an error response.",
		)
		return
	}

	if httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(motmedelHttpTypes.HttpContext); ok {
		httpContext.Response = &http.Response{
			StatusCode: responseWriter.WrittenStatusCode,
			Header:     responseWriter.Header(),
		}
		httpContext.ResponseBody = responseWriter.WrittenBody
	}
}
