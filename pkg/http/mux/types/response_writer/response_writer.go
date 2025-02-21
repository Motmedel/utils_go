package response_writer

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
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

func (responseWriter *ResponseWriter) WriteResponse(response *muxTypesResponse.Response) error {
	if response == nil {
		return nil
	}

	if responseWriter == nil {
		return motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrNilResponseWriter)
	}

	defaultHeaders := responseWriter.DefaultHeaders
	if defaultHeaders == nil {
		defaultHeaders = DefaultHeaders
	}

	defaultDocumentHeaders := responseWriter.DefaultHeaders
	if defaultDocumentHeaders == nil {
		defaultDocumentHeaders = DefaultDocumentHeaders
	}

	skippedDefaultHeadersSet := make(map[string]struct{})

	body := response.Body
	bodyStreamer := response.BodyStreamer

	var contentTypeString *string

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

	if response.StatusCode != 0 {
		responseWriter.WriteHeader(response.StatusCode)
	}

	if bodyStreamer != nil {
		flusher, ok := responseWriter.ResponseWriter.(http.Flusher)
		if !ok {
			return muxErrors.ErrNoResponseWriterFlusher
		}

		if _, ok := responseWriterHeader["transfer-encoding"]; ok {
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
