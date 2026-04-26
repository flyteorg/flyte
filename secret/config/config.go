package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "secret"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8093,
		Host: "0.0.0.0",
	},
	Kubernetes: KubernetesConfig{
		Namespace:   "flyte",
		ClusterName: "flyte-devbox",
		QPS:         100,
		Burst:       200,
		Timeout:     "30s",
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Secret service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Kubernetes configuration
	Kubernetes KubernetesConfig `json:"kubernetes"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	// Namespace where secrets are managed
	Namespace string `json:"namespace" pflag:",Kubernetes namespace for secret operations"`

	// ClusterName is the logical name of this cluster, used in secret status reporting
	ClusterName string `json:"clusterName" pflag:",Logical name of the cluster for secret status reporting"`

	// KubeConfig path (optional - if empty, uses in-cluster config)
	KubeConfig string `json:"kubeconfig" pflag:",Path to kubeconfig file (optional)"`

	// QPS is the maximum number of queries per second to the API server
	QPS int `json:"qps" pflag:",Maximum queries per second to the Kubernetes API server"`

	// Burst is the maximum burst above QPS allowed when talking to the API server
	Burst int `json:"burst" pflag:",Maximum burst above QPS for Kubernetes API server requests"`

	// Timeout is the request timeout for Kubernetes API server calls (e.g. "30s")
	Timeout string `json:"timeout" pflag:",Request timeout for Kubernetes API server calls"`
}

// GetConfig returns the parsed secret service configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
