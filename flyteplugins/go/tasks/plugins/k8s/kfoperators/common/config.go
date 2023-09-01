package common

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		Timeout: config.Duration{Duration: 1 * time.Minute},
	}

	configSection = pluginsConfig.MustRegisterSubSection("kf-operator", &defaultConfig)
)

// Config is config for 'pytorch' plugin
type Config struct {
	// If kubeflow operator doesn't update the status of the task after this timeout, the task will be considered failed.
	Timeout config.Duration `json:"timeout,omitempty"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
