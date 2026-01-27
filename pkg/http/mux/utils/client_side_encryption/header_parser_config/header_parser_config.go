package header_parser_config

import "github.com/go-jose/go-jose/v4"

const (
	DefaultHeaderName        = "X-Client-Public-Jwk"
	DefaultKeyAlgorithm      = "ECDH-ES"
	DefaultContentEncryption = "A256GCM"
)

var DefaultEncrypterOptions = (&jose.EncrypterOptions{}).WithContentType("application/json")

type Config struct {
	HeaderName        string
	KeyAlgorithm      jose.KeyAlgorithm
	ContentEncryption jose.ContentEncryption
	EncrypterOptions  *jose.EncrypterOptions
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		HeaderName:        DefaultHeaderName,
		KeyAlgorithm:      DefaultKeyAlgorithm,
		ContentEncryption: DefaultContentEncryption,
		EncrypterOptions:  DefaultEncrypterOptions,
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithHeaderName(headerName string) Option {
	return func(config *Config) {
		config.HeaderName = headerName
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

func WithEncrypterOptions(encrypterOptions *jose.EncrypterOptions) Option {
	return func(config *Config) {
		config.EncrypterOptions = encrypterOptions
	}
}
