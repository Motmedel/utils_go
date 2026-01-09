package fetch_config

import (
	"net/http"

	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config/retry_config"
)

type Option func(*Config)

const (
	DefaultMethod = "GET"
)

type Config struct {
	Method               string
	Headers              map[string]string
	Body                 []byte
	SkipReadResponseBody bool
	SkipErrorOnStatus    bool
	RetryConfig          *retry_config.Config
	HttpClient           *http.Client
}

func New(options ...Option) *Config {
	config := &Config{
		Method:     DefaultMethod,
		HttpClient: http.DefaultClient,
	}

	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	return config
}

func WithMethod(method string) Option {
	return func(configuration *Config) {
		configuration.Method = method
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(configuration *Config) {
		configuration.Headers = headers
	}
}

func WithBody(body []byte) Option {
	return func(configuration *Config) {
		configuration.Body = body
	}
}

func WithSkipReadResponseBody(skipReadResponseBody bool) Option {
	return func(configuration *Config) {
		configuration.SkipReadResponseBody = skipReadResponseBody
	}
}

func WithSkipErrorOnStatus(skipErrorOnStatus bool) Option {
	return func(configuration *Config) {
		configuration.SkipErrorOnStatus = skipErrorOnStatus
	}
}

func WithRetryConfig(retryConfig *retry_config.Config) Option {
	return func(configuration *Config) {
		configuration.RetryConfig = retryConfig
	}
}

func WithHttpClient(httpClient *http.Client) Option {
	return func(configuration *Config) {
		configuration.HttpClient = httpClient
	}
}
