package static_content

import (
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
)

type StaticContentData struct {
	Data         []byte
	Etag         string
	LastModified string
	Headers      []*muxTypesResponse.HeaderEntry
}

type StaticContent struct {
	StaticContentData
	ContentEncodingToData map[string]*StaticContentData
}
