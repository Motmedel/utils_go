package client_side_encryption

import (
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/request_parser/header_extractor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/client_side_encryption/body_parser_config"
	"github.com/Motmedel/utils_go/pkg/http/mux/utils/client_side_encryption/header_parser_config"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
	"github.com/Motmedel/utils_go/pkg/utils"
	"github.com/go-jose/go-jose/v4"
)

type HeaderParser struct {
	headerExtractor   *header_extractor.Parser
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
	EncrypterOptions  *jose.EncrypterOptions
}

func (p *HeaderParser) Parse(request *http.Request) (jose.Encrypter, *response_error.ResponseError) {
	clientJwkRaw, responseError := p.headerExtractor.Parse(request)
	if responseError != nil {
		return nil, responseError
	}

	var clientJwk jose.JSONWebKey
	if err := clientJwk.UnmarshalJSON([]byte(clientJwkRaw)); err != nil {
		return nil, &response_error.ResponseError{
			ClientError: motmedelErrors.NewWithTrace(
				fmt.Errorf("json web key unmarshal json: %w", err),
				clientJwkRaw,
			),
			ProblemDetail: problem_detail.New(
				http.StatusBadRequest,
				problem_detail_config.WithDetail(fmt.Sprintf("Invalid JWK header")),
			),
		}
	}

	clientJwkKey := clientJwk.Key
	if clientJwkKey == nil {
		return nil, &response_error.ResponseError{
			ProblemDetail: problem_detail.New(
				http.StatusBadRequest,
				problem_detail_config.WithDetail("Missing client public JWK key."),
			),
		}
	}

	clientJwkKeyId := clientJwk.KeyID
	responseEncrypter, err := jose.NewEncrypter(
		p.ContentEncryption,
		jose.Recipient{
			Algorithm: p.KeyAlgorithm,
			Key:       clientJwkKey,
			KeyID:     clientJwkKeyId,
		},
		p.EncrypterOptions,
	)
	if err != nil {
		return nil, &response_error.ResponseError{
			ClientError: motmedelErrors.NewWithTrace(
				fmt.Errorf("jose new encrypter: %w", err),
				clientJwkKey, clientJwkKeyId,
			),
			ProblemDetail: problem_detail.New(
				http.StatusBadRequest,
				problem_detail_config.WithDetail("Malformed JWK key."),
			),
		}
	}

	return responseEncrypter, nil
}

func NewHeaderParser(options ...header_parser_config.Option) (*HeaderParser, error) {
	config := header_parser_config.New(options...)

	headerExtractor, err := header_extractor.New(config.HeaderName)
	if err != nil {
		return nil, fmt.Errorf("header extractor new: %w", err)
	}

	return &HeaderParser{
		headerExtractor:   headerExtractor,
		KeyAlgorithm:      config.KeyAlgorithm,
		ContentEncryption: config.ContentEncryption,
		EncrypterOptions:  config.EncrypterOptions,
	}, nil
}

type BodyParser struct {
	PrivateKey        any
	KeyIdentifier     string
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
}

func (p *BodyParser) Parse(_ *http.Request, body []byte) ([]byte, *response_error.ResponseError) {
	jwe, err := jose.ParseEncrypted(
		string(body),
		[]jose.KeyAlgorithm{p.KeyAlgorithm},
		[]jose.ContentEncryption{p.ContentEncryption},
	)
	if err != nil {
		return nil, &response_error.ResponseError{
			ClientError: motmedelErrors.NewWithTrace(
				fmt.Errorf("jose parse encrypted: %w", err),
				body, p.KeyAlgorithm, p.ContentEncryption,
			),
		}
	}

	if keyIdentifier := p.KeyIdentifier; keyIdentifier != "" {
		jweKeyIdentifier := jwe.Header.KeyID
		if jweKeyIdentifier != keyIdentifier {
			return nil, &response_error.ResponseError{
				ProblemDetail: problem_detail.New(
					http.StatusBadRequest,
					problem_detail_config.WithDetail("The key identifier in the JWE header does not match the identifier of the key in use."),
					problem_detail_config.WithExtension(map[string]any{"kid": jweKeyIdentifier}),
				),
			}
		}
	}

	plaintext, err := jwe.Decrypt(p.PrivateKey)
	if err != nil {
		return nil, &response_error.ResponseError{
			ServerError: motmedelErrors.NewWithTrace(fmt.Errorf("jose web encryption decrypt: %w", err)),
		}
	}

	return plaintext, nil
}

func NewBodyParser(privateKey any, options ...body_parser_config.Option) (*BodyParser, error) {
	if utils.IsNil(privateKey) {
		return nil, motmedelErrors.NewWithTrace(nil_error.New("private key"))
	}

	config := body_parser_config.New(options...)

	return &BodyParser{
		PrivateKey:        privateKey,
		KeyAlgorithm:      config.KeyAlgorithm,
		ContentEncryption: config.ContentEncryption,
	}, nil
}
