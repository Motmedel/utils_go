package cookie_extractor_config

import "net/http"

var (
	DefaultProblemDetailStatusCode = http.StatusBadRequest
	DefaultProblemDetailText       = "Missing cookie."
)

type Config struct {
	ProblemDetailStatusCode int
	ProblemDetailText       string
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		ProblemDetailStatusCode: DefaultProblemDetailStatusCode,
		ProblemDetailText:       DefaultProblemDetailText,
	}
	for _, option := range options {
		option(config)
	}

	return config
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
