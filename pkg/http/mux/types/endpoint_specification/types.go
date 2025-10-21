package endpoint_specification

import (
	"net/http"

	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/request_parser"
	muxTypesParsing "github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	muxTypesRateLimiting "github.com/Motmedel/utils_go/pkg/http/mux/types/rate_limiting"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	httpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

type EndpointSpecification struct {
	Path                        string
	Method                      string
	Handler                     func(*http.Request, []byte) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError)
	StaticContent               *muxTypesStaticContent.StaticContent
	RateLimitingConfiguration   *muxTypesRateLimiting.RateLimitingConfiguration
	AuthenticationConfiguration *muxTypesParsing.AuthenticationConfiguration
	UrlParserConfiguration      *muxTypesParsing.UrlParserConfiguration
	HeaderParserConfiguration   *muxTypesParsing.HeaderParserConfiguration
	BodyParserConfiguration     *muxTypesParsing.BodyParserConfiguration
	CorsRequestParser           request_parser.RequestParser[*httpTypes.CorsConfiguration]
	DisableFetchMedata          bool
}
