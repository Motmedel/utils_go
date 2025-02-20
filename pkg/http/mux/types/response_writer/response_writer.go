package response_writer

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"net/http"
	"strings"
)

type ResponseWriter struct {
	http.ResponseWriter
	IsHeadRequest      bool
	WriteHeaderCalled  bool
	WriteCalled        bool
	NoStoreWrittenBody bool

	WrittenStatusCode int
	WrittenBody       []byte

	DefaultHeaders map[string]string
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

	// TODO: Rework default headers... Use some kind of tagging for responses? (check Chrome dev types -- see the extension I wrote for a list?)
	defaultHeaders := responseWriter.DefaultHeaders
	skippedDefaultHeadersSet := make(map[string]struct{})

	body := response.Body
	bodyStreamer := response.BodyStreamer

	responseWriterHeader := responseWriter.Header()
	for _, header := range response.Headers {
		if header == nil {
			continue
		}

		if strings.ToLower(header.Name) == "content-type" && len(body) == 0 && bodyStreamer == nil {
			continue
		}

		if _, ok := defaultHeaders[header.Name]; ok {
			if header.Overwrite {
				skippedDefaultHeadersSet[header.Name] = struct{}{}
			} else {
				continue
			}
		}

		responseWriterHeader.Set(header.Name, header.Value)
	}
	for headerName, headerValue := range defaultHeaders {
		if _, ok := skippedDefaultHeadersSet[headerName]; ok {
			continue
		}
		responseWriterHeader.Set(headerName, headerValue)
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
