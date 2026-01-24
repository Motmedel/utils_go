package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/authenticate_config"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/token/raw_token"
	"github.com/Motmedel/utils_go/pkg/utils"
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

func (token *Token) HeaderFields() map[string]any {
	return token.Header
}

func (token *Token) Claims() map[string]any {
	return token.Payload
}

func NewFromRawToken(rawToken *raw_token.Token) (*Token, error) {
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

	token, err := NewFromRawToken(rawToken)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("from raw token: %w", err), rawToken)
	}

	return token, nil
}

func Authenticate(tokenString string, options ...authenticate_config.Option) (*Token, error) {
	if tokenString == "" {
		return nil, nil
	}

	rawToken, err := raw_token.New(tokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: raw token new: %w", motmedelErrors.ErrParseError, err)
	}
	if rawToken == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrNilRawToken)
	}

	token, err := NewFromRawToken(rawToken)
	if err != nil {
		return nil, motmedelErrors.New(
			fmt.Errorf("%w: token from raw token: %w", motmedelErrors.ErrParseError, err),
			rawToken,
		)
	}

	config := authenticate_config.New(options...)

	signatureVerifier := config.SignatureVerifier
	if !utils.IsNil(signatureVerifier) {
		if token == nil {
			return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrNilToken)
		}

		tokenHeader := token.Header
		if tokenHeader == nil {
			return token, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrNilTokenHeader)
		}

		alg, err := utils.MapGetConvert[string](tokenHeader, "alg")
		if err != nil {
			var wrappedErr error = motmedelErrors.New(fmt.Errorf("map get convert: %w", err), tokenHeader)
			if motmedelErrors.IsAny(err, motmedelErrors.ErrConversionNotOk, motmedelErrors.ErrNotInMap) {
				wrappedErr = fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, wrappedErr)
			}
			return token, wrappedErr
		}

		verifierMethodName := signatureVerifier.GetName()
		if alg != verifierMethodName {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %w", motmedelErrors.ErrVerificationError, motmedelJwtErrors.ErrAlgorithmMismatch),
				motmedelJwtErrors.ErrAlgorithmMismatch, alg, verifierMethodName,
			)
		}

		if err := rawToken.Verify(signatureVerifier); err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("raw token verify: %w", err), rawToken)
		}
	} else if !config.AllowUnauthenticated {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	tokenValidator := config.TokenValidator
	if !utils.IsNil(tokenValidator) {
		// TODO: Should I really add `ErrValidationError` here?
		if err := tokenValidator.Validate(token); err != nil {
			return nil, motmedelErrors.New(
				fmt.Errorf("%w: token validator validate: %w", motmedelErrors.ErrValidationError, err),
				token,
			)
		}
	}

	return token, nil
}
