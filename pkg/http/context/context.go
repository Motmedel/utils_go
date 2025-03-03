package context

import (
	"context"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

type requestIdContextType struct{}

var RequestIdContextKey = &requestIdContextType{}

type httpContextContextType struct{}

var HttpContextContextKey httpContextContextType

type retryConfigurationContextType struct{}

var RetryConfigurationContextKey retryConfigurationContextType

func WithRetryConfiguration(parent context.Context, retryConfiguration *motmedelHttpTypes.RetryConfiguration) context.Context {
	return context.WithValue(parent, HttpContextContextKey, retryConfiguration)
}
