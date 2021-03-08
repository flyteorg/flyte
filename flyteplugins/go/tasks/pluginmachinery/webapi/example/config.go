package example

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		WebAPI: webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				Burst: 100,
				QPS:   10,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				Burst: 100,
				QPS:   10,
			},
			Caching: webapi.CachingConfig{
				Size:           500000,
				ResyncInterval: config.Duration{Duration: 30 * time.Second},
				Workers:        10,
			},
			ResourceMeta: nil,
		},

		ResourceConstraints: core.ResourceConstraintsSpec{
			ProjectScopeResourceConstraint: &core.ResourceConstraint{
				Value: 100,
			},
			NamespaceScopeResourceConstraint: &core.ResourceConstraint{
				Value: 50,
			},
		},
	}

	configSection = pluginsConfig.MustRegisterSubSection("admin", &defaultConfig)
)

// The config object for this plugin.
type Config struct {
	// Contains the default configs needed for the webapi base implementation.
	WebAPI              webapi.PluginConfig          `json:"webApi" pflag:",Defines config for the base WebAPI plugin."`
	ResourceConstraints core.ResourceConstraintsSpec `json:"resourceConstraints" pflag:"-,Defines resource constraints on how many executions to be created per project/overall at any given time."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
