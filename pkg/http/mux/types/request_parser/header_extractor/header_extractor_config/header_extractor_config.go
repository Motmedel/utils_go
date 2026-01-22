package header_extractor_config

import "net/http"

var (
	DefaultProblemDetailStatusCode   = http.StatusBadRequest
	DefaultProblemDetailMissingText  = "Missing header."
	DefaultProblemDetailMultipleText = "Multiple headers values."
)

type Config struct {
	ProblemDetailStatusCode   int
	ProblemDetailMissingText  string
	ProblemDetailMultipleText string
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		ProblemDetailStatusCode:   DefaultProblemDetailStatusCode,
		ProblemDetailMissingText:  DefaultProblemDetailMissingText,
		ProblemDetailMultipleText: DefaultProblemDetailMultipleText,
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
