package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
)

func MakeGzipData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	compression := gzip.BestCompression
	gzipWriter, err := gzip.NewWriterLevel(&buffer, compression)
	if err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("gzip new writer level: %w", err),
			compression,
		)
	}
	defer gzipWriter.Close()

	if _, err := gzipWriter.Write(data); err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("gzip writer write: %w", err))
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("gzip writer close: %w", err))
	}

	return buffer.Bytes(), nil
}
