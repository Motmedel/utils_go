package types

import (
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"iter"
	"net/http"
)

type HeaderEntry struct {
	Name      string
	Value     string
	Overwrite bool
}

type ResponseInfo struct {
	StatusCode   int
	Body         []byte
	BodyStreamer iter.Seq[[]byte]
	Headers      []*HeaderEntry
}

type HandlerErrorResponse struct {
	ClientError     error
	ServerError     error
	ProblemDetail   *problem_detail.ProblemDetail
	ResponseHeaders []*HeaderEntry
}

type HandlerSpecification struct {
	Path                      string
	Method                    string
	Handler                   func(*http.Request, []byte) (*ResponseInfo, *HandlerErrorResponse)
	StaticContent             *StaticContent
	RateLimitingConfiguration *RateLimitingConfiguration
	BodyParserConfiguration   *BodyParserConfiguration
}

type BodyParserConfiguration struct {
	ContentType string
	AllowEmpty  bool
	Parser      func(*http.Request, []byte) (any, *HandlerErrorResponse)
}
