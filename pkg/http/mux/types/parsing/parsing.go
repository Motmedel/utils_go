package parsing

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser"
	"net/http"
)

type parsedRequestUrlContextType struct{}
type parsedRequestHeaderContextType struct{}
type parsedRequestBodyContextType struct{}

var ParsedRequestUrlContextKey = parsedRequestUrlContextType{}
var ParsedRequestHeaderContextKey parsedRequestHeaderContextType
var ParsedRequestBodyContextKey parsedRequestBodyContextType

type UrlParserConfiguration struct {
	Parser func(*http.Request) (any, *response_error.ResponseError)
}

type HeaderParserConfiguration struct {
	Parser func(*http.Request) (any, *response_error.ResponseError)
}

type BodyParserConfiguration struct {
	ContentType string
	AllowEmpty  bool
	MaxBytes    int64
	Parser      body_parser.BodyParser
}
