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
		DefaultAgent: Agent{
			Endpoint:       "dns:///flyteagent.flyte.svc.cluster.local:80",
			Insecure:       true,
			DefaultTimeout: config.Duration{Duration: 10 * time.Second},
			KeepAliveParameters: &KeepAliveParameters{
				Time:                config.Duration{Duration: 10 * time.Second},
				Timeout:             config.Duration{Duration: 5 * time.Second},
				PermitWithoutStream: true,
			},
		},
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
	DefaultAgent Agent `json:"defaultAgent" pflag:",The default agent."`

	// The agents used to match against specific task types. {AgentId: Agent}
	Agents map[string]*Agent `json:"agents" pflag:",The agents."`

	// Maps task types to their agents. {TaskType: AgentId}
	AgentForTaskTypes map[string]string `json:"agentForTaskTypes" pflag:"-,"`
}

// KeepAliveParameters defines keepalive parameters on the client-side. For more details, check https://pkg.go.dev/google.golang.org/grpc/keepalive#ClientParameters
type KeepAliveParameters struct {
	// After a duration of this time if the client doesn't see any activity it
	// pings the server to see if the transport is still alive.
	// If set below 10s, a minimum value of 10s will be used instead.
	Time config.Duration `json:"time"`
	// After having pinged for keepalive check, the client waits for a duration
	// of Timeout and if no activity is seen even after that the connection is
	// closed.
	Timeout config.Duration `json:"timeout"`
	// If true, client sends keepalive pings even with no active RPCs. If false,
	// when there are no active RPCs, Time and Timeout will be ignored and no
	// keepalive pings will be sent.
	PermitWithoutStream bool `json:"permitWithoutStream"`
}

type Agent struct {
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

	// KeepAliveParameters defines keepalive parameters for the gRPC client
	KeepAliveParameters *KeepAliveParameters `json:"keepAliveParameters"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
