package endpoint

import (
	"net/http"
	"reflect"

	"github.com/Motmedel/utils_go/pkg/http/mux/types/body_loader"
	muxTypesRateLimiting "github.com/Motmedel/utils_go/pkg/http/mux/types/rate_limiting"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesStaticContent "github.com/Motmedel/utils_go/pkg/http/mux/types/static_content"
	httpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

type Hint struct {
	InputType         reflect.Type
	OutputType        reflect.Type
	OutputContentType string
	OutputOptional    bool
}

type Endpoint struct {
	Path                      string
	Method                    string
	RateLimitingConfiguration *muxTypesRateLimiting.RateLimitingConfiguration
	AuthenticationParser      request_parser.RequestParser[any]
	UrlParser                 request_parser.RequestParser[any]
	HeaderParser              request_parser.RequestParser[any]
	BodyLoader                *body_loader.Loader
	CorsParser                request_parser.RequestParser[*httpTypes.CorsConfiguration]
	DisableFetchMedata        bool
	Hint                      *Hint
	Handler                   func(*http.Request, []byte) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError)
	StaticContent             *muxTypesStaticContent.StaticContent
}
