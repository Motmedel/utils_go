package jwk_parse_config

import (
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	"github.com/Motmedel/utils_go/pkg/jwt/types/base_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/types/header_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/types/jwk_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/types/registered_claims_validator"
	"github.com/Motmedel/utils_go/pkg/jwt/types/tokenapi"
	"github.com/Motmedel/utils_go/pkg/jwt/types/validation_setting"
)

var DefaultValidator = &jwk_validator.JwkValidator{
	BaseValidator: base_validator.BaseValidator{
		HeaderValidator: &header_validator.HeaderValidator{
			Settings: map[string]validation_setting.Setting{
				"alg": validation_setting.Required,
			},
		},
		PayloadValidator: &registered_claims_validator.RegisteredClaimsValidator{
			Settings: map[string]validation_setting.Setting{
				"key": validation_setting.Required,
			},
		},
	},
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
