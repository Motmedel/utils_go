package parsing

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/request_parser"
)

type parsedRequestUrlContextType struct{}
type parsedRequestHeaderContextType struct{}
type parsedRequestBodyContextType struct{}

var ParsedRequestUrlContextKey = parsedRequestUrlContextType{}
var ParsedRequestHeaderContextKey parsedRequestHeaderContextType
var ParsedRequestBodyContextKey parsedRequestBodyContextType

type UrlParserConfiguration struct {
	Parser request_parser.RequestParser[any]
}

type HeaderParserConfiguration struct {
	Parser request_parser.RequestParser[any]
}

type BodyParserConfiguration struct {
	ContentType string
	AllowEmpty  bool
	MaxBytes    int64
	Parser      body_parser.BodyParser
}
