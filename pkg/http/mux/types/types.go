package types

import (
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"net/http"
)

type HeaderEntry struct {
	Name      string
	Value     string
	Overwrite bool
}

type ResponseInfo struct {
	StatusCode int
	Body       []byte
	Headers    []*HeaderEntry
}

type HandlerErrorResponse struct {
	ClientError     error
	ServerError     error
	ProblemDetail   *problem_detail.ProblemDetail
	ResponseHeaders []*HeaderEntry
}

type HandlerSpecification struct {
	Path                string
	Method              string
	ExpectedContentType string
	Handler             func(*http.Request, []byte) (*ResponseInfo, *HandlerErrorResponse)
	StaticContent       *StaticContent
}
