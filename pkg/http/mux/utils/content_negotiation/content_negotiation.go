package content_negotiation

import (
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpHeadersParsingAccept "github.com/Motmedel/utils_go/pkg/http/parsing/headers/accept"
	motmedelHttpHeadersParsingAcceptEncoding "github.com/Motmedel/utils_go/pkg/http/parsing/headers/accept_encoding"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
)

func GetContentNegotiation(requestHeader http.Header, strict bool) (*motmedelHttpTypes.ContentNegotiation, error) {
	if len(requestHeader) == 0 {
		return nil, nil
	}

	var contentNegotiation motmedelHttpTypes.ContentNegotiation

	if acceptValue := requestHeader.Get("Accept"); acceptValue != "" {
		acceptData := []byte(acceptValue)
		accept, err := motmedelHttpHeadersParsingAccept.ParseAccept(acceptData)
		if err != nil && strict {
			return nil, motmedelErrors.New(fmt.Errorf("parse accept: %w", err), acceptData)
		}
		contentNegotiation.Accept = accept
	}

	if acceptEncodingValue := requestHeader.Get("Accept-Encoding"); acceptEncodingValue != "" {
		acceptEncodingData := []byte(acceptEncodingValue)
		acceptEncoding, err := motmedelHttpHeadersParsingAcceptEncoding.ParseAcceptEncoding(acceptEncodingData)
		if err != nil && strict {
			return nil, motmedelErrors.New(fmt.Errorf("parse accept encoding: %w", err), acceptEncodingData)
		}
		contentNegotiation.AcceptEncoding = acceptEncoding
	}

	return &contentNegotiation, nil
}
