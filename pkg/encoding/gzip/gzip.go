package gzip

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"log/slog"
)

func MakeGzipData(ctx context.Context, data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	compression := gzip.BestCompression
	gzipWriter, err := gzip.NewWriterLevel(&buffer, compression)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("gzip new writer level: %w", err),
			compression,
		)
	}
	defer func() {
		if err := gzipWriter.Close(); err != nil {
			slog.WarnContext(
				motmedelContext.WithErrorContextValue(
					ctx,
					motmedelErrors.NewWithTrace(fmt.Errorf("gzip writer close: %w", err)),
				),
				"An error occurred when closing a gzip writer.",
			)
		}
	}()

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("gzip writer write: %w", err))
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("gzip writer close: %w", err))
	}

	return buffer.Bytes(), nil
}
