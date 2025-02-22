package response

import "iter"

type HeaderEntry struct {
	Name      string
	Value     string
	Overwrite bool
}

type Response struct {
	StatusCode    int
	Headers       []*HeaderEntry
	Body          []byte
	BodyStreamer  iter.Seq2[[]byte, error]
	SensitiveBody bool
}
