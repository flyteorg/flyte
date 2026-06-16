package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "manager"

//go:generate pflags Config --default-var=defaultConfig

// Config holds configuration for the unified Flyte Manager
type Config struct {
	// Server configuration - single port for all Connect services
	Server ServerConfig `json:"server"`

	// Executor configuration
	Executor ExecutorConfig `json:"executor"`

	// Kubernetes configuration for the shared client used by all services.
	Kubernetes app.K8sConfig `json:"kubernetes"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
}

// ExecutorConfig holds executor-specific configuration
type ExecutorConfig struct {
	HealthProbePort int `json:"healthProbePort" pflag:",Port for executor health probes"`
}

var defaultConfig = &Config{
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8090,
	},
	Executor: ExecutorConfig{
		HealthProbePort: 8081,
	},
	Kubernetes: app.K8sConfig{
		Namespace:  "flyte",
		KubeConfig: "",
		QPS:        1000,
		Burst:      2000,
		Timeout:    "30s",
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// GetConfig retrieves the current config value or default.
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
