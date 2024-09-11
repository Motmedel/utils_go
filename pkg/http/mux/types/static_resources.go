package types

type StaticContentData struct {
	Data         []byte
	Etag         string
	LastModified string
	Headers      []*HeaderEntry
}

type StaticContent struct {
	StaticContentData
	ContentEncodingToData map[string]*StaticContentData
}
