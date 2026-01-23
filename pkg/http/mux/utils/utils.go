package utils

import (
	"context"
	"fmt"

	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
)

func getParsed[T any](ctx context.Context, key any) (T, error) {
	value, err := motmedelContext.GetContextValue[T](ctx, key)
	if err != nil {
		return value, fmt.Errorf("get context value: %w", err)
	}

	return value, nil
}

func getNonZeroParsed[T comparable](ctx context.Context, key any) (T, error) {
	value, err := motmedelContext.GetNonZeroContextValue[T](ctx, key)
	if err != nil {
		return value, fmt.Errorf("get non zero context value: %w", err)
	}

	return value, nil
}

func GetServerNonZeroContextValue[T comparable](ctx context.Context, key any) (T, *response_error.ResponseError) {
	value, err := getNonZeroParsed[T](ctx, key)
	if err != nil {
		var zero T
		return zero, &response_error.ResponseError{
			ServerError: fmt.Errorf("get non zero context value: %w", err),
		}
	}

	return value, nil
}

func GetServerContextValue[T any](ctx context.Context, key any) (T, *response_error.ResponseError) {
	value, err := getParsed[T](ctx, key)
	if err != nil {
		var zero T
		return zero, &response_error.ResponseError{
			ServerError: fmt.Errorf("get context value: %w", err),
		}
	}

	return value, nil
}

func GetParsedRequestBody[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetServerParsedRequestBody[T any](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerContextValue[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetNonZeroParsedRequestBody[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetServerNonZeroParsedRequestBody[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestBodyContextKey)
}

func GetParsedRequestHeaders[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetNonZeroParsedRequestHeaders[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetServerNonZeroParsedRequestHeaders[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestHeaderContextKey)
}

func GetParsedRequestUrl[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func GetServerNonZeroParsedRequestUrl[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func GetNonZeroParsedRequestUrl[T comparable](ctx context.Context) (T, error) {
	return getNonZeroParsed[T](ctx, parsing.ParsedRequestUrlContextKey)
}

func GetParsedRequestAuthentication[T any](ctx context.Context) (T, error) {
	return getParsed[T](ctx, parsing.ParsedRequestAuthenticationContextKey)
}

func GetServerNonZeroParsedRequestAuthentication[T comparable](ctx context.Context) (T, *response_error.ResponseError) {
	return GetServerNonZeroContextValue[T](ctx, parsing.ParsedRequestAuthenticationContextKey)
}
