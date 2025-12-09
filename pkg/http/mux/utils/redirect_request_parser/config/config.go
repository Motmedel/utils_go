package config

type Option func(*Config)

const (
	DefaultParameterName = "redirect"
	DefaultRequireProto  = false
)

type Config struct {
	ParameterName string
	RequireProto  bool
}

func New(options ...Option) *Config {
	config := &Config{
		ParameterName: DefaultParameterName,
		RequireProto:  DefaultRequireProto,
	}

	for _, option := range options {
		if option != nil {
			option(config)
		}
	}

	return config
}

func WithParameterName(name string) Option {
	return func(config *Config) {
		config.ParameterName = name
	}
}

func WithRequireProto(requireProto bool) Option {
	return func(config *Config) {
		config.RequireProto = requireProto
	}
}
