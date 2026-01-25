package authenticator_with_key_handler_config

import (
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims/registered_claims"
)

type Config struct {
	ClaimsValidator validator.Validator[registered_claims.ParsedClaims]
	HeaderValidator validator.Validator[map[string]any]
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithClaimsValidator(claimsValidator validator.Validator[registered_claims.ParsedClaims]) Option {
	return func(config *Config) {
		config.ClaimsValidator = claimsValidator
	}
}

func WithHeaderValidator(headerValidator validator.Validator[map[string]any]) Option {
	return func(config *Config) {
		config.HeaderValidator = headerValidator
	}
}
