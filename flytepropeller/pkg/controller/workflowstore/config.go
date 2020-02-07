package workflowstore

import (
	ctrlConfig "github.com/lyft/flytepropeller/pkg/controller/config"
)

//go:generate pflags Config --default-var=defaultConfig

type Policy = string

const (
	PolicyInMemory             = "InMemory"
	PolicyPassThrough          = "PassThrough"
	PolicyResourceVersionCache = "ResourceVersionCache"
)

var (
	defaultConfig = &Config{
		Policy: PolicyPassThrough,
	}

	configSection = ctrlConfig.MustRegisterSubSection("workflowStore", defaultConfig)
)

type Config struct {
	Policy Policy `json:"policy" pflag:",Workflow Store Policy to initialize"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(*cfg)
}
