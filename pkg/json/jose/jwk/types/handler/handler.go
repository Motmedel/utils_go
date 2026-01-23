package handler

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/handler/handler_config"
	jwkKey "github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key"
	motmedelNetErrors "github.com/Motmedel/utils_go/pkg/net/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type Handler struct {
	JwkUrl *url.URL
	config *handler_config.Config

	keysMutex     sync.Mutex
	keys          []map[string]any
	keysExpiresAt *time.Time

	mu              sync.RWMutex
	keyIdToVerifier map[string]motmedelCryptoInterfaces.NamedVerifier
}

func (h *Handler) GetNamedVerifier(ctx context.Context, keyId string) (motmedelCryptoInterfaces.NamedVerifier, error) {
	h.keysMutex.Lock()
	err := func() error {
		defer h.keysMutex.Unlock()
		if expiresAt := h.keysExpiresAt; expiresAt == nil || expiresAt.Before(time.Now()) {
			jwkUrl := h.JwkUrl
			if jwkUrl == nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("%w (jwk url)", motmedelNetErrors.ErrNilUrl))
			}

			urlString := jwkUrl.String()
			response, keysResponseData, err := motmedelHttpUtils.FetchJson[jwkKey.Keys](ctx, urlString, h.config.FetchOptions...)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("fetch json: %w", err), urlString)
			}

			h.keys = keysResponseData.Keys

			responseHeader := response.Header
			expiresValue, err := motmedelHttpUtils.GetSingleHeader("Expires", responseHeader)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("get expires header: %w", err), responseHeader)
			}

			headerValueExpiresAt, err := time.Parse(time.RFC1123, expiresValue)
			if err != nil {
				return motmedelErrors.NewWithTrace(fmt.Errorf("time parse (expires): %w", err), expiresValue)
			}
			h.keysExpiresAt = &headerValueExpiresAt

			h.mu.Lock()
			clear(h.keyIdToVerifier)
			h.mu.Unlock()
		}

		return nil
	}()
	if err != nil {
		return nil, err
	}

	h.mu.RLock()
	if verifier, ok := h.keyIdToVerifier[keyId]; ok {
		h.mu.RUnlock()
		return verifier, nil
	}
	h.mu.RUnlock()

	keys := h.keys
	for _, keyMap := range keys {
		if keyMap == nil {
			continue
		}

		keyMapKeyId := keyMap["kid"]
		if keyMapKeyId != keyId {
			continue
		}

		key, err := jwkKey.New(keyMap)
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("new key: %w", err), keyMap)
		}
		if key == nil {
			return nil, motmedelErrors.NewWithTrace(motmedelJwkErrors.ErrNilKey)
		}

		namedVerifier, err := key.NamedVerifier()
		if err != nil {
			return nil, motmedelErrors.New(fmt.Errorf("key named verifier: %w", err), key)
		}
		if utils.IsNil(namedVerifier) {
			return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
		}

		h.mu.Lock()
		h.keyIdToVerifier[keyId] = namedVerifier
		h.mu.Unlock()

		return namedVerifier, nil
	}

	return nil, nil
}

func New(jwkUrl *url.URL, options ...handler_config.Option) (*Handler, error) {
	if jwkUrl == nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w (jwk url)", motmedelNetErrors.ErrNilUrl))
	}

	return &Handler{
		JwkUrl:          jwkUrl,
		keyIdToVerifier: make(map[string]motmedelCryptoInterfaces.NamedVerifier),
		config:          handler_config.New(options...),
	}, nil
}
