package database

import (
	"context"
	"time"
)

var (
	DefaultTimeout = 5 * time.Second
)

func MakeTimeoutCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, DefaultTimeout)
}
