package dask

import (
	pluginsConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/logs"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		Logs: logs.DefaultConfig,
	}

	configSection = pluginsConfig.MustRegisterSubSection("dask", &defaultConfig)
)

// Config is config for 'dask' plugin
type Config struct {
	Logs logs.LogConfig `json:"logs,omitempty"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
