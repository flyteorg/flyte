package workflowstore

import (
	ctrlConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
)

//go:generate pflags Config --default-var=defaultConfig

type Policy = string

const (
	// PolicyInMemory provides an inmemory Workflow store which is useful for testing
	PolicyInMemory = "InMemory"
	// PolicyPassThrough just calls the underlying Clientset or the shared informer cache to get or write the workflow
	PolicyPassThrough = "PassThrough"
	// PolicyTrackTerminated tracks terminated workflows
	PolicyTrackTerminated = "TrackTerminated"
	// PolicyResourceVersionCache uses the resource version on the Workflow object, to determine if the inmemory copy
	// of the workflow is stale
	PolicyResourceVersionCache = "ResourceVersionCache"
)

// By default we will use the ResourceVersionCache example
var (
	defaultConfig = &Config{
		Policy: PolicyResourceVersionCache,
	}

	configSection = ctrlConfig.MustRegisterSubSection("workflowStore", defaultConfig)
)

// Config for Workflow access in the controller.
// Various policies are available like - InMemory, PassThrough, TrackTerminated, ResourceVersionCache
type Config struct {
	Policy Policy `json:"policy" pflag:",Workflow Store Policy to initialize"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(*cfg)
}
