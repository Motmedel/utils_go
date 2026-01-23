package authenticator

import (
	"context"
	"fmt"

	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwkErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwk/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/handler"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/authenticate_config"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/authenticator/authenticator_config"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/authenticator/authenticator_with_key_handler_config"
	motmedelJwkToken "github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/token"
	motmedelJwtValidator "github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/validator"
	motmedelMaps "github.com/Motmedel/utils_go/pkg/maps"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type Authenticator struct {
	config *authenticator_config.Config
}

func (a *Authenticator) Authenticate(ctx context.Context, tokenString string) (*motmedelJwkToken.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	signatureVerifier := a.config.SignatureVerifier
	validator := &motmedelJwtValidator.Validator{
		HeaderValidator:  a.config.HeaderValidator,
		PayloadValidator: a.config.ClaimsValidator,
	}
	token, err := motmedelJwkToken.Authenticate(
		tokenString,
		authenticate_config.WithSignatureVerifier(signatureVerifier),
		authenticate_config.WithTokenValidator(validator),
	)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("token authenticate: %w", err),
			tokenString, signatureVerifier, validator,
		)
	}
	if token == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrNilToken)
	}

	return token, nil
}

func New(options ...authenticator_config.Option) *Authenticator {
	return &Authenticator{
		config: authenticator_config.New(options...),
	}
}

type AuthenticatorWithKeyHandler struct {
	Handler *handler.Handler
	config  *authenticator_with_key_handler_config.Config
}

func (a *AuthenticatorWithKeyHandler) Authenticate(ctx context.Context, tokenString string) (*motmedelJwkToken.Token, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context err: %w", err)
	}

	token, err := motmedelJwkToken.New(tokenString)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w: token new %w", motmedelErrors.ErrParseError, err))
	}
	if token == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrNilToken)
	}

	tokenHeader := token.Header
	if tokenHeader == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelJwtErrors.ErrNilTokenHeader)
	}

	kid, err := motmedelMaps.MapGetConvert[string](tokenHeader, "kid")
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("map get convert (kid): %w", err), tokenHeader)
	}

	keyHandler := a.Handler
	if keyHandler == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelJwkErrors.ErrNilHandler)
	}
	signatureVerifier, err := keyHandler.GetNamedVerifier(ctx, kid)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("handler get named verifier: %w", err), kid)
	}
	if utils.IsNil(signatureVerifier) {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	validator := &motmedelJwtValidator.Validator{
		HeaderValidator:  a.config.HeaderValidator,
		PayloadValidator: a.config.ClaimsValidator,
	}

	parsedToken, err := motmedelJwkToken.Authenticate(
		tokenString,
		authenticate_config.WithSignatureVerifier(signatureVerifier),
		authenticate_config.WithTokenValidator(validator),
	)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("token authenticate: %w", err),
			tokenString, signatureVerifier, validator,
		)
	}
	if parsedToken == nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w (parsed)", motmedelJwtErrors.ErrNilToken))
	}

	return parsedToken, nil
}

func NewWithKeyHandler(handler *handler.Handler, options ...authenticator_with_key_handler_config.Option) (*AuthenticatorWithKeyHandler, error) {
	if handler == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelJwkErrors.ErrNilHandler)
	}
	return &AuthenticatorWithKeyHandler{
		Handler: handler,
		config:  authenticator_with_key_handler_config.New(options...),
	}, nil
}
