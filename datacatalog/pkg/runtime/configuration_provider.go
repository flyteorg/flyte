package runtime

// Interface for getting parsed values from a configuration file
type Configuration interface {
	ApplicationConfiguration() ApplicationConfiguration
}

// Implementation of a Configuration
type ConfigurationProvider struct {
	applicationConfiguration ApplicationConfiguration
}

func (p *ConfigurationProvider) ApplicationConfiguration() ApplicationConfiguration {
	return p.applicationConfiguration
}

func NewConfigurationProvider() Configuration {
	return &ConfigurationProvider{
		applicationConfiguration: NewApplicationConfigurationProvider(),
	}
}
