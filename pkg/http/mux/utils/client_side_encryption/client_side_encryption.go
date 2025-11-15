package client_side_encryption

import (
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"github.com/go-jose/go-jose/v4"
)

type HeaderRequestParser struct {
	Header            string
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
	EncrypterOptions  *jose.EncrypterOptions
}

func (parser *HeaderRequestParser) Parse(request *http.Request) (jose.Encrypter, *response_error.ResponseError) {
	if request == nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequest),
		}
	}

	requestHeader := request.Header
	if requestHeader == nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(motmedelHttpErrors.ErrNilHttpRequestHeader),
		}
	}

	_, ok := requestHeader[parser.Header]
	if !ok {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				fmt.Sprintf("Missing %q header.", parser.Header),
				nil,
			),
		}
	}
	clientJwkRaw := requestHeader.Get(parser.Header)
	if clientJwkRaw == "" {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				fmt.Sprintf("Empty %q header.", parser.Header),
				nil,
			),
		}
	}

	var clientJwk jose.JSONWebKey
	if err := clientJwk.UnmarshalJSON([]byte(clientJwkRaw)); err != nil {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				fmt.Sprintf("Invalid %q header.", parser.Header),
				nil,
			),
			ClientError: motmedelErrors.NewWithTrace(
				fmt.Errorf("json web key unmarshal json: %w", err),
				clientJwkRaw,
			),
		}
	}
	clientJwkKey := clientJwk.Key
	if clientJwkKey == nil {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				"Missing client public JWK key.",
				nil,
			),
		}
	}

	clientJwkKeyId := clientJwk.KeyID
	responseEncrypter, err := jose.NewEncrypter(
		parser.ContentEncryption,
		jose.Recipient{
			Algorithm: parser.KeyAlgorithm,
			Key:       clientJwkKey,
			KeyID:     clientJwkKeyId,
		},
		parser.EncrypterOptions,
	)
	if err != nil {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.MakeBadRequestProblemDetail(
				"Malformed client public JWK key.",
				nil,
			),
			ClientError: motmedelErrors.NewWithTrace(
				fmt.Errorf("jose new encrypter: %w", err),
				clientJwkKey, clientJwkKeyId,
			),
		}
	}

	return responseEncrypter, nil
}

type BodyParser struct {
	PrivateKey        any
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
}

func (bodyParser *BodyParser) Parse(_ *http.Request, body []byte) ([]byte, *response_error.ResponseError) {
	jwe, err := jose.ParseEncrypted(
		string(body),
		[]jose.KeyAlgorithm{bodyParser.KeyAlgorithm},
		[]jose.ContentEncryption{bodyParser.ContentEncryption},
	)
	if err != nil {
		return nil, &response_error.ResponseError{
			ClientError: motmedelErrors.NewWithTrace(
				fmt.Errorf("jose parse encrypted: %w", err),
				body, bodyParser.KeyAlgorithm, bodyParser.ContentEncryption,
			),
		}
	}

	plaintext, err := jwe.Decrypt(bodyParser.PrivateKey)
	if err != nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(fmt.Errorf("jose web encryption decrypt: %w", err)),
		}
	}

	return plaintext, nil
}
