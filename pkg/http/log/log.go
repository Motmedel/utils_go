package log

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Motmedel/ecs_go/ecs"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	"log/slog"
)

var DefaultHeaderExtractor = ecs.DefaultMaskedHeaderExtractor

type ContextHandler struct {
	slog.Handler
	headerExtractor func(requestResponse any) string
}

func (contextHandler *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext); ok {
		headerExtractor := contextHandler.headerExtractor
		if headerExtractor == nil {
			headerExtractor = DefaultHeaderExtractor
		}

		base, err := ecs.ParseHttpContext(httpContext, headerExtractor)
		if err != nil {
			return motmedelErrors.MakeError(
				fmt.Errorf("ecs parse http context: %w", err),
				httpContext,
			)
		}

		baseBytes, err := json.Marshal(base)
		if err != nil {
			return motmedelErrors.MakeErrorWithStackTrace(
				fmt.Errorf("json marshal (http context ecs base): %w", err),
				base,
			)
		}

		var baseMap map[string]any
		if err = json.Unmarshal(baseBytes, &baseMap); err != nil {
			return motmedelErrors.MakeErrorWithStackTrace(
				fmt.Errorf("json unmarshal (http context ecs base map): %w", err),
				baseMap,
			)
		}

		record.Add(motmedelLog.AttrsFromMap(baseMap)...)
	}

	return contextHandler.Handler.Handle(ctx, record)
}
