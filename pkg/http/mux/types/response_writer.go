package types

import "net/http"

type ResponseWriter struct {
	http.ResponseWriter
	IsHeadRequest     bool
	WriteHeaderCaller bool
	WriteCalled       bool

	WrittenStatusCode   int
	WrittenResponseBody []byte

	DefaultHeaders map[string]string
}

func (responseWriter *ResponseWriter) WriteHeader(statusCode int) {
	responseWriter.WriteHeaderCaller = true
	responseWriter.WrittenStatusCode = statusCode
	responseWriter.ResponseWriter.WriteHeader(statusCode)
}

func (responseWriter *ResponseWriter) Write(data []byte) (int, error) {
	responseWriter.WriteCalled = true

	if !responseWriter.WriteHeaderCaller {
		statusCode := http.StatusOK
		if len(data) == 0 {
			statusCode = http.StatusNoContent
		}
		responseWriter.WriteHeader(statusCode)
	}

	if responseWriter.IsHeadRequest {
		return 0, nil
	}

	responseWriter.WrittenResponseBody = data
	return responseWriter.ResponseWriter.Write(data)
}
