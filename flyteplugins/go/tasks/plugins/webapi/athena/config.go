package athena

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/config"
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

		DefaultWorkGroup: "primary",
		DefaultCatalog:   "AwsDataCatalog",
	}

	configSection = pluginsConfig.MustRegisterSubSection("athena", &defaultConfig)
)

type Config struct {
	WebAPI              webapi.PluginConfig          `json:"webApi" pflag:",Defines config for the base WebAPI plugin."`
	ResourceConstraints core.ResourceConstraintsSpec `json:"resourceConstraints" pflag:"-,Defines resource constraints on how many executions to be created per project/overall at any given time."`
	DefaultWorkGroup    string                       `json:"defaultWorkGroup" pflag:",Defines the default workgroup to use when running on Athena unless overwritten by the task."`
	DefaultCatalog      string                       `json:"defaultCatalog" pflag:",Defines the default catalog to use when running on Athena unless overwritten by the task."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
