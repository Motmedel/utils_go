package parsing

import (
	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_parser"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser"
)

type EmptyOption int

const (
	BodyRequired EmptyOption = iota
	BodyOptional
	BodyForbidden
)

type parsedRequestUrlContextType struct{}
type parsedRequestHeaderContextType struct{}
type parsedRequestBodyContextType struct{}
type parsedRequestAuthenticationContextType struct{}

var ParsedRequestUrlContextKey = parsedRequestUrlContextType{}
var ParsedRequestHeaderContextKey = parsedRequestHeaderContextType{}
var ParsedRequestBodyContextKey = parsedRequestBodyContextType{}
var ParsedRequestAuthenticationContextKey = parsedRequestAuthenticationContextType{}

type UrlParserConfiguration struct {
	Parser request_parser.RequestParser[any]
}

type HeaderParserConfiguration struct {
	Parser request_parser.RequestParser[any]
}

type AuthenticationConfiguration struct {
	Parser request_parser.RequestParser[any]
}

type BodyParserConfiguration struct {
	ContentType string
	EmptyOption EmptyOption
	MaxBytes    int64
	Parser      body_parser.BodyParser[any]
}
