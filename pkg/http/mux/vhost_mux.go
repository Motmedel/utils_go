package mux

import (
	"crypto/tls"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxInternalVhostMux "github.com/Motmedel/utils_go/pkg/http/mux/internal/vhost_mux"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	muxTypesResponseError "github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	muxTypesResponseWriter "github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"net"
	"net/http"
)

type VhostMuxSpecification struct {
	Mux         http.Handler
	RedirectTo  string
	Certificate *tls.Certificate
}

type VhostMux struct {
	baseMux
	HostToSpecification map[string]*VhostMuxSpecification
}

func (vhostMux *VhostMux) PatchHttpServer(httpServer *http.Server) {
	if httpServer == nil {
		return
	}

	httpServer.Handler = vhostMux

	tlsConfig := httpServer.TLSConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
		httpServer.TLSConfig = tlsConfig
	}

	tlsConfig.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if clientHello == nil {
			return nil, nil
		}

		hostToSpecification := vhostMux.HostToSpecification
		if hostToSpecification == nil {
			return nil, motmedelErrors.NewWithTrace(muxErrors.ErrNilHostToMuxSpecification)
		}

		specification, ok := hostToSpecification[clientHello.ServerName]
		if !ok || specification == nil {
			return nil, nil
		}

		return specification.Certificate, nil
	}
}

func vhostMuxHandleRequest(
	vhostMux *VhostMux,
	request *http.Request,
	responseWriter http.ResponseWriter,
) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
	if vhostMux == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilVhostMux),
		}
	}

	if request == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	host, _, err := net.SplitHostPort(request.Host)
	if err != nil {
		host = request.Host
	}

	hostToSpecification := vhostMux.HostToSpecification
	if hostToSpecification == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilHostToMuxSpecification),
		}
	}

	muxSpecification, ok := hostToSpecification[host]
	if !ok {
		return nil, &muxTypesResponseError.ResponseError{
			ProblemDetail: problem_detail.MakeStatusCodeProblemDetail(
				http.StatusMisdirectedRequest,
				"",
				nil,
			),
		}
	}
	if muxSpecification == nil {
		return nil, &muxTypesResponseError.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilMuxSpecification),
		}
	}

	if redirectTo := muxSpecification.RedirectTo; redirectTo != "" {
		return &muxTypesResponse.Response{
			StatusCode: http.StatusMovedPermanently,
			Headers: []*muxTypesResponse.HeaderEntry{
				{Name: "Location", Value: muxInternalVhostMux.HexEscapeNonASCII(redirectTo + request.RequestURI)},
			},
		}, nil
	} else if muxSpecificationMux := muxSpecification.Mux; muxSpecificationMux != nil {
		muxSpecificationMux.ServeHTTP(responseWriter, request)
		return nil, nil
	}

	return nil, &muxTypesResponseError.ResponseError{
		ServerError: motmedelErrors.NewWithTrace(
			muxErrors.ErrUnusableMuxSpecification,
			muxSpecification,
		),
	}
}

func (vhostMux *VhostMux) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	vhostMux.baseMux.ServeHttpWithCallback(
		responseWriter,
		request,
		func(request *http.Request, responseWriter *muxTypesResponseWriter.ResponseWriter) (*muxTypesResponse.Response, *muxTypesResponseError.ResponseError) {
			response, responseError := vhostMuxHandleRequest(vhostMux, request, responseWriter)
			if responseError != nil {
				responseError.ProblemDetailConverter = vhostMux.ProblemDetailConverter
			}

			if responseWriter != nil {
				responseWriter.DefaultHeaders = vhostMux.DefaultHeaders
				responseWriter.DefaultDocumentHeaders = vhostMux.DefaultDocumentHeaders
			}

			return response, responseError
		},
	)
}
