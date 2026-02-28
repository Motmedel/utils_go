package http_context_extractor

import (
	"context"
	"fmt"
	"log/slog"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelJson "github.com/Motmedel/utils_go/pkg/json"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	schemaUtils "github.com/Motmedel/utils_go/pkg/schema/utils"
)

type Extractor struct {
}

func (e *Extractor) Handle(ctx context.Context, record *slog.Record) error {
	if record == nil {
		return nil
	}

	if requestId, ok := ctx.Value(motmedelHttpContext.RequestIdContextKey).(string); ok {
		record.Add(slog.Group("http", slog.Group("request", slog.String("id", requestId))))
	}

	if httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext); ok {
		base, err := schemaUtils.ParseHttpContext(httpContext)
		if err != nil {
			return motmedelErrors.New(
				fmt.Errorf("ecs parse http context: %w", err),
				httpContext,
			)
		}
		if base != nil {
			baseMap, err := motmedelJson.ObjectToMap(base)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("object to map: %w", err), base)
			}

			record.Add(motmedelLog.AttrsFromMap(baseMap)...)
		}
	}

	return nil
}

func New() *Extractor {
	return &Extractor{}
}
