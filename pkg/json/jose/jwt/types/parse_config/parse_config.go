package parse_config

import (
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/base_validator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/registered_claims_validator"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/tokenapi"
)

var DefaultValidator = &base_validator.BaseValidator{
	PayloadValidator: &registered_claims_validator.RegisteredClaimsValidator{},
}

type Config struct {
	SignatureVerifier motmedelCryptoInterfaces.NamedVerifier
	TokenValidator    validator.Validator[tokenapi.Token]
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		TokenValidator: DefaultValidator,
	}

	for _, option := range options {
		option(config)
	}
	return config
}

func WithSignatureVerifier(signatureVerifier motmedelCryptoInterfaces.NamedVerifier) Option {
	return func(configuration *Config) {
		configuration.SignatureVerifier = signatureVerifier
	}
}

func WithTokenValidator(tokenValidator validator.Validator[tokenapi.Token]) Option {
	return func(configuration *Config) {
		configuration.TokenValidator = tokenValidator
	}
}
