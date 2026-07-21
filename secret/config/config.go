package config

import (
	webhookconfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "secret"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8093,
		Host: "0.0.0.0",
	},
	Kubernetes: app.K8sConfig{
		// Same constant the pod webhook's embedded K8s secret fetcher defaults to —
		// the webhook reads back the Secrets this service writes.
		Namespace:   webhookconfig.DefaultSecretsNamespace,
		QPS:         100,
		Burst:       200,
		Timeout:     "30s",
		ClusterName: "flyte-devbox",
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Secret service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Kubernetes client configuration (also carries ClusterName for status reporting).
	Kubernetes app.K8sConfig `json:"kubernetes"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed secret service configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
