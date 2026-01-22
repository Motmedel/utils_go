package token_cookie_extractor_config

import "net/http"

var (
	DefaultCookieName              = "Authorization"
	DefaultCookiePrefix            = "Bearer "
	DefaultProblemDetailStatusCode = http.StatusUnauthorized
	DefaultProblemDetailText       = "Missing token cookie."
)

type Config struct {
	CookieName              string
	CookiePrefix            string
	ProblemDetailStatusCode int
	ProblemDetailText       string
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{
		CookieName:              DefaultCookieName,
		CookiePrefix:            DefaultCookiePrefix,
		ProblemDetailStatusCode: DefaultProblemDetailStatusCode,
		ProblemDetailText:       DefaultProblemDetailText,
	}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithCookieName(cookieName string) Option {
	return func(configuration *Config) {
		configuration.CookieName = cookieName
	}
}

func WithCookiePrefix(cookiePrefix string) Option {
	return func(configuration *Config) {
		configuration.CookiePrefix = cookiePrefix
	}
}

func WithProblemDetailStatusCode(statusCode int) Option {
	return func(configuration *Config) {
		configuration.ProblemDetailStatusCode = statusCode
	}
}

func WithProblemDetailText(text string) Option {
	return func(configuration *Config) {
		configuration.ProblemDetailText = text
	}
}
