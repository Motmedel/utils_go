package body_parser_config

import "github.com/go-jose/go-jose/v4"

const (
	DefaultKeyAlgorithm      = "ECDH-ES"
	DefaultContentEncryption = "A256GCM"
)

type Config struct {
	KeyIdentifier     string
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		KeyAlgorithm:      DefaultKeyAlgorithm,
		ContentEncryption: DefaultContentEncryption,
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithKeyIdentifier(keyIdentifier string) Option {
	return func(config *Config) {
		config.KeyIdentifier = keyIdentifier
	}
}

func WithKeyAlgorithm(keyAlgorithm jose.KeyAlgorithm) Option {
	return func(config *Config) {
		config.KeyAlgorithm = keyAlgorithm
	}
}

func WithContentEncryption(contentEncryption jose.ContentEncryption) Option {
	return func(config *Config) {
		config.ContentEncryption = contentEncryption
	}
}
