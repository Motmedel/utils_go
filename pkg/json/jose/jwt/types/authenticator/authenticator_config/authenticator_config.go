package authenticator_config

import (
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims/parsed_claims"
)

type Config struct {
	SignatureVerifier motmedelCryptoInterfaces.NamedVerifier
	ClaimsValidator   validator.Validator[parsed_claims.Claims]
	HeaderValidator   validator.Validator[map[string]any]
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithSignatureVerifier(signatureVerifier motmedelCryptoInterfaces.NamedVerifier) Option {
	return func(config *Config) {
		config.SignatureVerifier = signatureVerifier
	}
}

func WithClaimsValidator(claimsValidator validator.Validator[parsed_claims.Claims]) Option {
	return func(config *Config) {
		config.ClaimsValidator = claimsValidator
	}
}

func WithHeaderValidator(headerValidator validator.Validator[map[string]any]) Option {
	return func(config *Config) {
		config.HeaderValidator = headerValidator
	}
}
