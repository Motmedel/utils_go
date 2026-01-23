package query_extractor_config

type Config struct {
	AllowAdditionalParameters bool
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithAllowAdditionalParameters(allowAdditionalParameters bool) Option {
	return func(config *Config) {
		config.AllowAdditionalParameters = allowAdditionalParameters
	}
}
