package retry_config

import (
	"net/http"
	"time"

	"github.com/Motmedel/utils_go/pkg/http/types/fetch_config/retry_config/response_checker"
)

var defaultResponseChecker = response_checker.New(
	func(response *http.Response, err error) bool {
		if response != nil {
			return response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
		}

		if err != nil {
			return true
		}

		return false
	},
)

type Option func(*Config)

const (
	DefaultCount     = 2
	DefaultBaseDelay = time.Duration(500) * time.Millisecond
)

type Config struct {
	Count           int
	BaseDelay       time.Duration
	MaximumWaitTime time.Duration
	ResponseChecker response_checker.ResponseChecker
}

func NewConfig(options ...Option) *Config {
	config := &Config{
		Count:           DefaultCount,
		BaseDelay:       DefaultBaseDelay,
		ResponseChecker: defaultResponseChecker,
	}

	for _, option := range options {
		option(config)
	}

	return config
}

func WithCount(count int) Option {
	return func(configuration *Config) {
		configuration.Count = count
	}
}

func WithBaseDelay(baseDelay time.Duration) Option {
	return func(configuration *Config) {
		configuration.BaseDelay = baseDelay
	}
}

func WithMaximumWaitTime(maximumWaitTime time.Duration) Option {
	return func(configuration *Config) {
		configuration.MaximumWaitTime = maximumWaitTime
	}
}

func WithResponseChecker(checker response_checker.ResponseChecker) Option {
	return func(configuration *Config) {
		configuration.ResponseChecker = checker
	}
}
