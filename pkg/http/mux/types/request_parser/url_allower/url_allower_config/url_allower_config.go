package url_processor_config

type Option func(*Config)

type Config struct {
	AllowLocalhost           bool
	AllowedDomains           []string
	AllowedRegisteredDomains []string
}

func New(options ...Option) *Config {
	config := &Config{}
	for _, option := range options {
		option(config)
	}

	return config
}

func WithAllowLocalhost(allowLocalhost bool) Option {
	return func(config *Config) {
		config.AllowLocalhost = allowLocalhost
	}
}

func WithAllowedDomains(allowedDomains []string) Option {
	return func(config *Config) {
		config.AllowedDomains = allowedDomains
	}
}

func WithAllowedRegisteredDomains(allowedRegisteredDomains []string) Option {
	return func(config *Config) {
		config.AllowedRegisteredDomains = allowedRegisteredDomains
	}
}
