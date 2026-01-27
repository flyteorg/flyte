package config

import "github.com/flyteorg/flyte/v2/flytestdlib/config"

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
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// ExecutorConfig holds executor-specific configuration
type ExecutorConfig struct {
	HealthProbePort int `json:"healthProbePort"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	Namespace  string `json:"namespace"`
	KubeConfig string `json:"kubeconfig"` // Optional, defaults to in-cluster or ~/.kube/config
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
	},
}

var cfg Config

// GetConfig retrieves the current config value or default.
func GetConfig() *Config {
	return &cfg
}

func init() {
	config.MustRegisterSection(configSectionKey, &cfg)
}
