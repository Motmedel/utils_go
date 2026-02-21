package http_context_extractor_config

type Config struct {
	ProjectId string
}

type Option func(*Config)

func New(options ...Option) *Config {
	config := &Config{}
	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	return config
}

func WithProjectId(projectId string) Option {
	return func(config *Config) {
		config.ProjectId = projectId
	}
}
