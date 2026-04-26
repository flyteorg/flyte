package config

import (
	"time"

	appconfig "github.com/flyteorg/flyte/v2/app/config"
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

	// Kubernetes configuration
	Kubernetes KubernetesConfig `json:"kubernetes"`

	// Apps is the control plane AppService configuration.
	Apps appconfig.AppConfig `json:"apps"`

	// InternalApps is the data plane InternalAppService configuration.
	InternalApps appconfig.InternalAppConfig `json:"internalApps"`
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

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	Namespace  string `json:"namespace" pflag:",Kubernetes namespace"`
	KubeConfig string `json:"kubeconfig" pflag:",Path to kubeconfig file (optional)"`
	QPS        int    `json:"qps" pflag:",Max sustained queries per second to the API server"`
	Burst      int    `json:"burst" pflag:",Max burst queries to the API server"`
	Timeout    string `json:"timeout" pflag:",Default timeout for API server requests (e.g. 30s)"`
}

var defaultConfig = &Config{
	Server: ServerConfig{
		Host: "0.0.0.0",
		Port: 8090,
	},
	Executor: ExecutorConfig{
		HealthProbePort: 8081,
	},
	Kubernetes: KubernetesConfig{
		Namespace:  "flyte",
		KubeConfig: "",
		QPS:        1000,
		Burst:      2000,
		Timeout:    "30s",
	},
	Apps: appconfig.AppConfig{
		CacheTTL: 30 * time.Second,
	},
	InternalApps: appconfig.InternalAppConfig{
		Enabled:               false,
		DefaultRequestTimeout: 300 * time.Second,
		MaxRequestTimeout:     3600 * time.Second,
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// GetConfig retrieves the current config value or default.
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
