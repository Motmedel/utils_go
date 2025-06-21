package log

import (
	"context"
	"fmt"
	"github.com/Motmedel/ecs_go/ecs"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelJson "github.com/Motmedel/utils_go/pkg/json"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	"log/slog"
)

var DefaultHeaderExtractor = ecs.DefaultMaskedHeaderExtractor

type HttpContextExtractor struct {
	HeaderExtractor func(any) string
}

func (httpContextExtractor *HttpContextExtractor) Handle(ctx context.Context, record *slog.Record) error {
	if record == nil {
		return nil
	}

	if requestId, ok := ctx.Value(motmedelHttpContext.RequestIdContextKey).(string); ok {
		record.Add(slog.Group("http", slog.Group("request", slog.String("id", requestId))))
	}

	if httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext); ok {
		headerExtractor := httpContextExtractor.HeaderExtractor
		if headerExtractor == nil {
			headerExtractor = DefaultHeaderExtractor
		}

		base, err := ecs.ParseHttpContext(httpContext, headerExtractor)
		if err != nil {
			return motmedelErrors.New(
				fmt.Errorf("ecs parse http context: %w", err),
				httpContext,
			)
		}

		baseMap, err := motmedelJson.ObjectToMap(base)
		if err != nil {
			return motmedelErrors.New(fmt.Errorf("object to map: %w", err), base)
		}

		record.Add(motmedelLog.AttrsFromMap(baseMap)...)
	}

	return nil
}
