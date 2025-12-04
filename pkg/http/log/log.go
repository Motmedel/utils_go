package log

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Motmedel/ecs_go/ecs"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpContext "github.com/Motmedel/utils_go/pkg/http/context"
	configPkg "github.com/Motmedel/utils_go/pkg/http/log/types/config"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelJson "github.com/Motmedel/utils_go/pkg/json"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
)

type HttpContextExtractor struct {
	HeaderExtractor func(any) string
}

func (extractor *HttpContextExtractor) Handle(ctx context.Context, record *slog.Record) error {
	if record == nil {
		return nil
	}

	if requestId, ok := ctx.Value(motmedelHttpContext.RequestIdContextKey).(string); ok {
		record.Add(slog.Group("http", slog.Group("request", slog.String("id", requestId))))
	}

	if httpContext, ok := ctx.Value(motmedelHttpContext.HttpContextContextKey).(*motmedelHttpTypes.HttpContext); ok {
		base, err := ecs.ParseHttpContext(httpContext, extractor.HeaderExtractor)
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

func New(options ...configPkg.Option) *HttpContextExtractor {
	config := configPkg.New(options...)
	return &HttpContextExtractor{HeaderExtractor: config.HeaderExtractor}
}
