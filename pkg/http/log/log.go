package log

import (
	"context"
	"log/slog"
)

type ContextHandler struct {
	slog.Handler
}

func (contextHandler *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	return contextHandler.Handler.Handle(ctx, record)
}
