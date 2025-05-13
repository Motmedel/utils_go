package log

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Motmedel/ecs_go/ecs"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	motmedelTlsContext "github.com/Motmedel/utils_go/pkg/tls/context"
	motmedelTlsTypes "github.com/Motmedel/utils_go/pkg/tls/types"
	"log/slog"
)

func ParseTlsContext(tlsContext *motmedelTlsTypes.TlsContext) *ecs.Base {
	if tlsContext == nil {
		return nil
	}

	connectionState := tlsContext.ConnectionState
	if connectionState == nil {
		return nil
	}

	var base ecs.Base
	ecs.EnrichWithTlsContext(&base, tlsContext)

	return &base
}

func ExtractTlsContext(ctx context.Context, record *slog.Record) error {
	if dnsContext, ok := ctx.Value(motmedelTlsContext.TlsContextKey).(*motmedelTlsTypes.TlsContext); ok && dnsContext != nil {
		base := ParseTlsContext(dnsContext)
		if base != nil {
			baseBytes, err := json.Marshal(base)
			if err != nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (ecs base): %w", err), base)
			}

			var baseMap map[string]any
			if err = json.Unmarshal(baseBytes, &baseMap); err != nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal (ecs base map): %w", err), baseMap)
			}

			record.Add(motmedelLog.AttrsFromMap(baseMap)...)
		}
	}

	return nil
}

var TlsContextExtractor = motmedelLog.ContextExtractorFunction(ExtractTlsContext)
