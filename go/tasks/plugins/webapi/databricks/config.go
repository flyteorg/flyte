package databricks

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/config"
)

var (
	defaultCluster = "COMPUTE_CLUSTER"
	tokenKey       = "FLYTE_DATABRICKS_API_TOKEN" // nolint: gosec

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
		DefaultCluster: defaultCluster,
		TokenKey:       tokenKey,
	}

	configSection = pluginsConfig.MustRegisterSubSection("databricks", &defaultConfig)
)

// Config is config for 'databricks' plugin
type Config struct {
	// WebAPI defines config for the base WebAPI plugin
	WebAPI webapi.PluginConfig `json:"webApi" pflag:",Defines config for the base WebAPI plugin."`

	// ResourceConstraints defines resource constraints on how many executions to be created per project/overall at any given time
	ResourceConstraints core.ResourceConstraintsSpec `json:"resourceConstraints" pflag:"-,Defines resource constraints on how many executions to be created per project/overall at any given time."`

	DefaultCluster string `json:"defaultWarehouse" pflag:",Defines the default warehouse to use when running on Databricks unless overwritten by the task."`

	TokenKey string `json:"databricksTokenKey" pflag:",Name of the key where to find Databricks token in the secret manager."`

	DatabricksInstance string `json:"databricksInstance" pflag:",Databricks workspace instance name."`

	EntrypointFile string `json:"entrypointFile" pflag:",A URL of the entrypoint file. DBFS and cloud storage (s3://, gcs://, adls://, etc) locations are supported."`
	// databricksEndpoint overrides databricks instance endpoint, only for testing
	databricksEndpoint string
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
