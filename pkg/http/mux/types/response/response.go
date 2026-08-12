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
	// InlineScriptHashes holds Content Security Policy hash sources (e.g.
	// "sha256-<base64>") of inline scripts occurring in the body, to be merged
	// into the effective script-src directive when responding.
	InlineScriptHashes []string
}
