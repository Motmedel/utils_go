package token_cookie_extractor_config

import "net/http"

var (
	DefaultName                    = "session"
	DefaultProblemDetailStatusCode = http.StatusUnauthorized
	DefaultProblemDetailText       = "Missing cookie with token."
)

type Config struct {
	Name                    string
	ProblemDetailStatusCode int
	ProblemDetailText       string
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		Name:                    DefaultName,
		ProblemDetailStatusCode: DefaultProblemDetailStatusCode,
		ProblemDetailText:       DefaultProblemDetailText,
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithName(name string) Option {
	return func(config *Config) {
		config.Name = name
	}
}

func WithProblemDetailStatusCode(statusCode int) Option {
	return func(config *Config) {
		config.ProblemDetailStatusCode = statusCode
	}
}

func WithProblemDetailText(text string) Option {
	return func(config *Config) {
		config.ProblemDetailText = text
	}
}
