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
	CacheInvalidationURL: "http://localhost:9444",
	Kubernetes: app.K8sConfig{
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

	// CacheInvalidationURL is the base URL of the pod webhook's cache invalidation server. After
	// a secret is written, the service posts there so pods admitted next pick up the new value
	// instead of waiting out the webhook's cache TTL. Defaults to localhost, which is correct
	// while the webhook shares this pod; point it at the webhook Service when they are split.
	// Empty disables invalidation (writes then take effect after the TTL).
	CacheInvalidationURL string `json:"cacheInvalidationURL" pflag:",Base URL of the pod webhook's secret cache invalidation server"`

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
