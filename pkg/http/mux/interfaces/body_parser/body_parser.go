package body_parser

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"net/http"
)

type BodyParser interface {
	Parse(*http.Request, []byte) (any, *response_error.ResponseError)
}

type BodyParserFunction func(*http.Request, []byte) (any, *response_error.ResponseError)

func (bpf BodyParserFunction) Parse(request *http.Request, body []byte) (any, *response_error.ResponseError) {
	return bpf(request, body)
}
