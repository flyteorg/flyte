package agent

import (
	"time"

	pluginsConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

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
		DefaultAgent: Deployment{
			Endpoint:       "",
			Insecure:       true,
			DefaultTimeout: config.Duration{Duration: 10 * time.Second},
		},
		// AsyncPlugin should be registered to at least one task type.
		// Reference: https://github.com/flyteorg/flyte/blob/master/flyteplugins/go/tasks/pluginmachinery/registry.go#L27
		SupportedTaskTypes: []string{"task_type_1", "task_type_2"},
		PollInterval:       config.Duration{Duration: 10 * time.Second},
	}

	configSection = pluginsConfig.MustRegisterSubSection("agent-service", &defaultConfig)
)

// Config is config for 'agent' plugin
type Config struct {
	// WebAPI defines config for the base WebAPI plugin
	WebAPI webapi.PluginConfig `json:"webApi" pflag:",Defines config for the base WebAPI plugin."`

	// ResourceConstraints defines resource constraints on how many executions to be created per project/overall at any given time
	ResourceConstraints core.ResourceConstraintsSpec `json:"resourceConstraints" pflag:"-,Defines resource constraints on how many executions to be created per project/overall at any given time."`

	// The default agent if there does not exist a more specific matching against task types
	DefaultAgent Deployment `json:"defaultAgent" pflag:",The default agent."`

	// The agents used to match against specific task types. {agentDeploymentID: AgentDeployment}
	AgentDeployments map[string]*Deployment `json:"agents" pflag:",The agents."`

	// Maps task types to their agents. {TaskType: agentDeploymentID}
	AgentForTaskTypes map[string]string `json:"agentForTaskTypes" pflag:"-,"`

	// SupportedTaskTypes is a list of task types that are supported by this plugin.
	SupportedTaskTypes []string `json:"supportedTaskTypes" pflag:"-,Defines a list of task types that are supported by this plugin."`

	// PollInterval is the interval at which the plugin should poll the agent for metadata updates
	PollInterval config.Duration `json:"pollInterval" pflag:",The interval at which the plugin should poll the agent for metadata updates."`
}

type Deployment struct {
	// Endpoint points to an agent gRPC endpoint
	Endpoint string `json:"endpoint"`

	// Insecure indicates whether the communication with the gRPC service is insecure
	Insecure bool `json:"insecure"`

	// DefaultServiceConfig sets default gRPC service config; check https://github.com/grpc/grpc/blob/master/doc/service_config.md for more details
	DefaultServiceConfig string `json:"defaultServiceConfig"`

	// Timeouts defines various RPC timeout values for different plugin operations: CreateTask, GetTask, DeleteTask; if not configured, defaults to DefaultTimeout
	Timeouts map[string]config.Duration `json:"timeouts"`

	// DefaultTimeout gives the default RPC timeout if a more specific one is not defined in Timeouts; if neither DefaultTimeout nor Timeouts is defined for an operation, RPC timeout will not be enforced
	DefaultTimeout config.Duration `json:"defaultTimeout"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
