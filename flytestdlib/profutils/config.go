package profutils

import "github.com/flyteorg/flyte/v2/flytestdlib/config"

//go:generate pflags Config --default-var=defaultConfig

const configSectionKey = "prof"

var (
	configSection = config.MustRegisterSection(configSectionKey, defaultConfig)
	defaultConfig = &Config{}
)

type Config struct {
	DisableConfigEndpoint bool `config:"DisableConfigEndpoint"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
