package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "queue"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8089,
		Host: "0.0.0.0",
	},
	Kubernetes: KubernetesConfig{
		Namespace: "flyte",
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Queue service
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
	// Namespace where TaskAction CRs will be created
	Namespace string `json:"namespace" pflag:",Kubernetes namespace for TaskAction CRs"`

	// KubeConfig path (optional - if empty, uses in-cluster config)
	KubeConfig string `json:"kubeconfig" pflag:",Path to kubeconfig file (optional)"`
}

// GetConfig returns the parsed queue configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
