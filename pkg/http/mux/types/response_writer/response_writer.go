package response_writer

import (
	"context"
	"fmt"
	motmedelGzip "github.com/Motmedel/utils_go/pkg/encoding/gzip"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	motmedelIter "github.com/Motmedel/utils_go/pkg/iter"
	"log/slog"
	"maps"
	"net/http"
	"strings"
)

var DefaultHeaders = map[string]string{
	"Cache-Control":                "no-store",
	"X-Content-Type-Options":       "nosniff",
	"Cross-Origin-Resource-Policy": "same-origin",
}

var DefaultDocumentHeaders = map[string]string{
	"Cross-Origin-Opener-Policy":   "same-origin",
	"Cross-Origin-Embedder-Policy": "require-corp",
	"Content-Security-Policy":      "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'",
	"Permissions-Policy":           "geolocation=(), microphone=(), camera=()",
}

type ResponseWriter struct {
	http.ResponseWriter
	IsHeadRequest      bool
	WriteHeaderCalled  bool
	WriteCalled        bool
	NoStoreWrittenBody bool

	WrittenStatusCode int
	WrittenBody       []byte

	DefaultHeaders         map[string]string
	DefaultDocumentHeaders map[string]string
}

func (responseWriter *ResponseWriter) WriteHeader(statusCode int) {
	responseWriter.WriteHeaderCalled = true
	responseWriter.WrittenStatusCode = statusCode
	responseWriter.ResponseWriter.WriteHeader(statusCode)
}

func (responseWriter *ResponseWriter) Write(data []byte) (int, error) {
	responseWriter.WriteCalled = true

	if !responseWriter.WriteHeaderCalled {
		statusCode := http.StatusOK
		if len(data) == 0 {
			statusCode = http.StatusNoContent
		}
		responseWriter.WriteHeader(statusCode)
	}

	if responseWriter.IsHeadRequest || len(data) == 0 {
		return 0, nil
	}

	if !responseWriter.NoStoreWrittenBody {
		responseWriter.WrittenBody = append(responseWriter.WrittenBody, data...)
	}

	n, err := responseWriter.ResponseWriter.Write(data)
	if err != nil {
		return n, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("http response writer write: %w", err),
		)
	}

	return n, nil
}

func (responseWriter *ResponseWriter) WriteResponse(
	ctx context.Context,
	response *muxTypesResponse.Response,
	acceptEncoding *motmedelHttpTypes.AcceptEncoding,
) error {
	if response == nil {
		return nil
	}

	var defaultHeaders map[string]string
	if responseWriterDefaultHeaders := responseWriter.DefaultHeaders; responseWriterDefaultHeaders == nil {
		defaultHeaders = DefaultHeaders
	} else {
		defaultHeaders = maps.Collect(
			motmedelIter.Concat2(
				maps.All(DefaultHeaders),
				maps.All(responseWriterDefaultHeaders),
			),
		)
	}

	var defaultDocumentHeaders map[string]string
	if responseWriterDefaultDocumentHeaders := responseWriter.DefaultDocumentHeaders; responseWriterDefaultDocumentHeaders == nil {
		defaultDocumentHeaders = DefaultDocumentHeaders
	} else {
		defaultDocumentHeaders = maps.Collect(
			motmedelIter.Concat2(
				maps.All(DefaultDocumentHeaders),
				maps.All(responseWriterDefaultDocumentHeaders),
			),
		)
	}

	skippedDefaultHeadersSet := make(map[string]struct{})

	body := response.Body
	bodyStreamer := response.BodyStreamer

	var contentTypeString *string
	var contentEncodingString *string
	var cacheControlString string

	responseWriterHeader := responseWriter.Header()
	for _, header := range response.Headers {
		if header == nil || header.Name == "" {
			continue
		}

		canonicalHeaderName := http.CanonicalHeaderKey(header.Name)
		headerValue := header.Value

		if canonicalHeaderName == "Content-Type" {
			contentTypeString = &headerValue
			if len(body) == 0 && bodyStreamer == nil {
				continue
			}
		}

		if canonicalHeaderName == "Content-Encoding" {
			contentEncodingString = &headerValue
			if len(body) == 0 && bodyStreamer == nil {
				continue
			}
		}

		if canonicalHeaderName == "Cache-Control" {
			cacheControlString = headerValue
		}

		if _, ok := defaultHeaders[canonicalHeaderName]; ok {
			if !header.Overwrite {
				continue
			}
			skippedDefaultHeadersSet[canonicalHeaderName] = struct{}{}
		}

		if _, ok := defaultDocumentHeaders[canonicalHeaderName]; ok {
			if !header.Overwrite {
				continue
			}
			skippedDefaultHeadersSet[canonicalHeaderName] = struct{}{}
		}

		responseWriterHeader.Set(canonicalHeaderName, headerValue)
	}
	for headerName, headerValue := range defaultHeaders {
		if _, ok := skippedDefaultHeadersSet[headerName]; ok {
			continue
		}
		responseWriterHeader.Set(headerName, headerValue)
	}

	if contentTypeString != nil {
		contentTypeData := []byte(*contentTypeString)
		contentType, err := content_type.ParseContentType(contentTypeData)
		if err != nil {
			return motmedelErrors.MakeError(fmt.Errorf("parse content type: %w", err), contentTypeData)
		}
		if contentType == nil {
			return motmedelErrors.MakeErrorWithStackTrace(content_type.ErrNilContentType, contentTypeData)
		}

		var useDocumentHeaders bool

		effectiveContentTypeValues := []string{
			strings.ToLower(contentType.Subtype),
			contentType.GetStructuredSyntaxName(true),
		}
		for _, effectiveContentTypeValue := range effectiveContentTypeValues {
			switch effectiveContentTypeValue {
			case "html", "xhtml", "xml", "svg":
				useDocumentHeaders = true
			}
		}

		if useDocumentHeaders {
			for headerName, headerValue := range defaultDocumentHeaders {
				if _, ok := skippedDefaultHeadersSet[headerName]; ok {
					continue
				}
				responseWriterHeader.Set(headerName, headerValue)
			}
		}
	}

	// Try to compress the body if it is of a decent size, and
	shouldTryToCompressBody := len(body) > 1000 &&
		// ... no content encoding is already applied
		contentEncodingString == nil &&
		// ... the client indicates that it supports encoded content
		acceptEncoding != nil &&
		// ... the response body is not sensitive (compressing could theoretically enable attacks)
		!response.SensitiveBody &&
		// ... the response concerns a non-static resource (static resources should provide encoded values explicitly,
		// and I don't want to add a `Vary` header like this)
		cacheControlString == "no-store"

	if shouldTryToCompressBody {
		// NOTE: The case where `identify` effectively has a quality value of 0 should be handled elsewhere.
		switch motmedelHttpUtils.GetMatchingContentEncoding(acceptEncoding.GetPriorityOrderedEncodings(), []string{"gzip"}) {
		case "gzip":
			gzipBody, err := motmedelGzip.MakeGzipData(body)
			if err != nil {
				slog.Default().WarnContext(
					context.WithValue(
						ctx,
						motmedelErrors.ErrorContextKey,
						motmedelErrors.MakeError(fmt.Errorf("make gzip data: %w", err), body),
					),
					"An error occurred when making Gzip data.",
				)
			}

			if len(gzipBody) < len(body) {
				body = gzipBody
				responseWriterHeader.Set("Content-Encoding", "gzip")
			}
		}
	}

	if response.StatusCode != 0 {
		responseWriter.WriteHeader(response.StatusCode)
	}

	if bodyStreamer != nil {
		flusher, ok := responseWriter.ResponseWriter.(http.Flusher)
		if !ok {
			return muxErrors.ErrNoResponseWriterFlusher
		}

		if _, ok := responseWriterHeader["Transfer-Encoding"]; ok {
			return muxErrors.ErrTransferEncodingAlreadySet
		}

		// TODO: Figure out how to support HTTP/2?
		responseWriterHeader.Set("Transfer-Encoding", "chunked")

		for bodyChunk, err := range bodyStreamer {
			if err != nil {
				return fmt.Errorf("body streamer: %w", err)
			}

			if _, err := responseWriter.Write(bodyChunk); err != nil {
				return fmt.Errorf("mux response writer write: %w", err)
			}
			flusher.Flush()
		}

		if _, err := responseWriter.Write([]byte{}); err != nil {
			return fmt.Errorf("mux response writer write (empty chunk): %w", err)
		}
	} else {
		if _, err := responseWriter.Write(body); err != nil {
			return fmt.Errorf("mux response writer write: %w", err)
		}
	}

	return nil
}
