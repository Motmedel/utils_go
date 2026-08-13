// Package brotli wraps the Brotli implementation used for compression, keeping
// the rest of the module independent of the underlying library.
//
// The implementation is github.com/andybalholm/brotli, the established pure-Go
// port of the reference implementation: a compression codec is correctness
// critical and maintained upstream, so it is wrapped rather than rewritten.
// Its one extra module requirement is test-only and does not reach builds.
package brotli

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/andybalholm/brotli"

	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

func MakeBrotliData(ctx context.Context, data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	quality := brotli.BestCompression
	brotliWriter := brotli.NewWriterLevel(&buffer, quality)

	// NOTE: Unlike a gzip writer, closing a brotli writer twice reports an
	// error, so the writer is only closed by the deferred function on early
	// returns.
	closed := false
	defer func() {
		if closed {
			return
		}
		if err := brotliWriter.Close(); err != nil {
			slog.WarnContext(
				motmedelContext.WithError(
					ctx,
					motmedelErrors.NewWithTrace(fmt.Errorf("brotli writer close: %w", err)),
				),
				"An error occurred when closing a brotli writer.",
			)
		}
	}()

	if _, err := brotliWriter.Write(data); err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("brotli writer write: %w", err),
			quality,
		)
	}

	closed = true
	if err := brotliWriter.Close(); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("brotli writer close: %w", err))
	}

	return buffer.Bytes(), nil
}
