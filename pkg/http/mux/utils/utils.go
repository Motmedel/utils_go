package utils

import (
	"context"
	"fmt"
	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
)

func getParsed[T any](ctx context.Context, key any) (T, error) {
	var zero T

	if err := ctx.Err(); err != nil {
		return zero, err
	}

	value, err := motmedelContext.GetContextValue[T](ctx, key)
	if err != nil {
		return zero, fmt.Errorf("get context value: %w", err)
	}

	return value, nil
}

func getNonZeroParsed[T comparable](ctx context.Context, key any) (T, error) {
	var zero T

	if err := ctx.Err(); err != nil {
		return zero, err
	}

	value, err := motmedelContext.GetNonZeroContextValue[T](ctx, key)
	if err != nil {
		return zero, fmt.Errorf("get non zero context value: %w", err)
	}

	return value, nil
}

func GetParsedRequestBody[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetServerParsedRequestBody[T any](ctx context.Context) (T, *response_error.ResponseError) {
	value, err := GetParsedRequestBody[T](ctx)
	if err != nil {
		var zero T
		return zero, &response_error.ResponseError{
			ServerError: fmt.Errorf("get parsed request body: %w", err),
		}
	}

	return value, nil
}

func GetNonZeroParsedRequestBody[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetServerNonZeroParsedRequestBody[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	value, err := GetNonZeroParsedRequestBody[T](ctx)
	if err != nil {
		var zero T
		return zero, &response_error.ResponseError{
			ServerError: fmt.Errorf("get non zero parsed request body: %w", err),
		}
	}

	return value, nil
}

func GetParsedRequestHeaders[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetNonZeroParsedRequestHeaders[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetParsedRequestUrl[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func GetNoneZeroParsedRequestUrl[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestUrlContextKey)
}
