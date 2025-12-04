package config

import (
	"github.com/Motmedel/ecs_go/ecs"
)

var DefaultHeaderExtractor = ecs.DefaultMaskedHeaderExtractor

type Option func(configuration *Config)

type Config struct {
	// TODO: Rework this. Use an interface instead?
	HeaderExtractor func(any) string
}

func New(options ...Option) *Config {
	config := &Config{
		HeaderExtractor: DefaultHeaderExtractor,
	}

	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	return config
}

func WithHeaderExtractor(headerExtractor func(any) string) Option {
	return func(configuration *Config) {
		configuration.HeaderExtractor = headerExtractor
	}
}
