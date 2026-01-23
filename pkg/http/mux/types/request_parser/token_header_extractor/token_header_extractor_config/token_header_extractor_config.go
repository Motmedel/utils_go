package token_header_extractor_config

import "net/http"

var (
	DefaultHeaderName                = "Authorization"
	DefaultHeaderValuePrefix         = "Bearer "
	DefaultProblemDetailStatusCode   = http.StatusUnauthorized
	DefaultProblemDetailMissingText  = "Missing token header."
	DefaultProblemDetailMultipleText = "Multiple token headers values."
)

type Config struct {
	HeaderName                string
	HeaderValuePrefix         string
	ProblemDetailStatusCode   int
	ProblemDetailMissingText  string
	ProblemDetailMultipleText string
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		HeaderName:                DefaultHeaderName,
		HeaderValuePrefix:         DefaultHeaderValuePrefix,
		ProblemDetailStatusCode:   DefaultProblemDetailStatusCode,
		ProblemDetailMissingText:  DefaultProblemDetailMissingText,
		ProblemDetailMultipleText: DefaultProblemDetailMultipleText,
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

func WithHeaderValuePrefix(valuePrefix string) Option {
	return func(config *Config) {
		config.HeaderValuePrefix = valuePrefix
	}
}

func WithProblemDetailStatusCode(statusCode int) Option {
	return func(config *Config) {
		config.ProblemDetailStatusCode = statusCode
	}
}

func WithProblemDetailMissingText(text string) Option {
	return func(config *Config) {
		config.ProblemDetailMissingText = text
	}
}

func WithProblemDetailMultipleText(text string) Option {
	return func(config *Config) {
		config.ProblemDetailMultipleText = text
	}
}
