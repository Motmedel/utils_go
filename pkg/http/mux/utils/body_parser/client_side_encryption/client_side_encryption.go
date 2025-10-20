package client_side_encryption

import (
	"fmt"
	"net/http"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_processor"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	"github.com/Motmedel/utils_go/pkg/utils"
	"github.com/go-jose/go-jose/v4"
)

type CseBodyParser struct {
	PrivateKey        any
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
}

func (bodyParser *CseBodyParser) Parse(_ *http.Request, body []byte) ([]byte, *response_error.ResponseError) {
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

type CseBodyParserWithProcessor[T any] struct {
	CseBodyParser
	Processor body_processor.BodyProcessor[T, []byte]
}

func (bodyParser *CseBodyParserWithProcessor[T]) Parse(request *http.Request, body []byte) (T, *response_error.ResponseError) {
	var zero T

	parser := bodyParser.CseBodyParser
	if utils.IsNil(parser) {
		return zero, &response_error.ResponseError{ServerError: muxErrors.ErrNilBodyParser}
	}

	result, responseError := parser.Parse(request, body)
	if responseError != nil {
		return zero, responseError
	}

	processor := bodyParser.Processor
	if utils.IsNil(processor) {
		return zero, &response_error.ResponseError{ServerError: motmedelErrors.NewWithTrace(muxErrors.ErrNilProcessor)}
	}

	processedResult, responseError := processor.Process(result)
	if responseError != nil {
		return zero, responseError
	}

	return processedResult, nil
}
