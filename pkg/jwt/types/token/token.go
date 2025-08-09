package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/raw_token"
	"github.com/Motmedel/utils_go/pkg/utils"
	"maps"
	"strings"
)

type Token struct {
	Header  map[string]any
	Payload map[string]any
}

func (token *Token) Encode(signer motmedelCryptoInterfaces.NamedSigner) (string, error) {
	if utils.IsNil(signer) {
		return "", motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilSigner)
	}

	payloadBytes, err := json.Marshal(token.Payload)
	if err != nil {
		return "", motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (payload): %w", err), token.Payload)
	}

	var header map[string]any
	if tokenHeader := token.Header; tokenHeader != nil {
		header = maps.Clone(tokenHeader)
		if header == nil {
			return "", motmedelErrors.NewWithTrace(fmt.Errorf("%w (header clone)", motmedelErrors.ErrNilMap))
		}
	} else {
		header = make(map[string]any)
		header["typ"] = "JWT"
	}

	header["alg"] = signer.GetName()

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (header): %w", err), header)
	}

	headerBase64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	signatureInput := []byte(strings.Join([]string{headerBase64, payloadBase64}, "."))

	signature, err := signer.Sign(signatureInput)
	if err != nil {
		return "", motmedelErrors.NewWithTrace(fmt.Errorf("signer sign: %w", err), signatureInput)
	}

	return strings.Join(
		[]string{headerBase64, payloadBase64, base64.RawURLEncoding.EncodeToString(signature)},
		".",
	), nil
}

func FromRawToken(rawToken *raw_token.RawToken) (*Token, error) {
	if rawToken == nil {
		return nil, nil
	}

	var token Token

	if err := json.Unmarshal(rawToken.Header, &token.Header); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal (header): %w", err), rawToken.Header)
	}

	if err := json.Unmarshal(rawToken.Payload, &token.Payload); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal (payload): %w", err), rawToken.Payload)
	}

	return &token, nil
}

func New(tokenString string) (*Token, error) {
	rawToken, err := raw_token.New(tokenString)
	if err != nil {
		return nil, fmt.Errorf("raw token new: %w", err)
	}

	token, err := FromRawToken(rawToken)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("from raw token: %w", err), rawToken)
	}

	return token, nil
}
