// Package bigquery implements WebAPI plugin for Google BigQuery
package bigquery

import (
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/google"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

var (
	defaultConfig = Config{
		WebAPI: webapi.PluginConfig{
			ResourceQuotas: map[core.ResourceNamespace]int{
				"default": 1000,
			},
			ReadRateLimiter: webapi.RateLimiterConfig{
				Burst: 100,
				QPS:   10,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				Burst: 100,
				QPS:   10,
			},
			Caching: webapi.CachingConfig{
				Size:              500000,
				ResyncInterval:    config.Duration{Duration: 30 * time.Second},
				Workers:           10,
				MaxSystemFailures: 5,
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
		GoogleTokenSource: google.GetDefaultConfig(),
	}

	configSection = pluginsConfig.MustRegisterSubSection("bigquery", &defaultConfig)
)

// Config is config for 'bigquery' plugin
type Config struct {
	// WebAPI defines config for the base WebAPI plugin
	WebAPI webapi.PluginConfig `json:"webApi" pflag:",Defines config for the base WebAPI plugin."`

	// ResourceConstraints defines resource constraints on how many executions to be created per project/overall at any given time
	ResourceConstraints core.ResourceConstraintsSpec `json:"resourceConstraints" pflag:"-,Defines resource constraints on how many executions to be created per project/overall at any given time."`

	// GoogleTokenSource configures token source for BigQuery client
	GoogleTokenSource google.TokenSourceFactoryConfig `json:"googleTokenSource" pflag:",Defines Google token source"`

	// bigQueryEndpoint overrides BigQuery client endpoint, only for testing
	bigQueryEndpoint string
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
