package authenticator_with_key_handler_config

import (
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims/parsed_claims"
)

type Config struct {
	ClaimsValidator validator.Validator[parsed_claims.Claims]
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
