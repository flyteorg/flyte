package config

import (
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "actions"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8091,
		Host: "0.0.0.0",
	},
	Kubernetes: KubernetesConfig{
		Namespace: "flyte",
	},
	WatchBufferSize: 100,
	WatchWorkers:    10,
	RunServiceURL:   "http://localhost:8090",
	// 8M slots × 8 bytes/pointer = 64 MB; can track ~8M unique actions.
	RecordFilterSize: 1 << 23,
	Apps: AppConfig{
		Enabled:               false,
		Namespace:             "flyte-apps",
		DefaultRequestTimeout: 5 * time.Minute,
		MaxRequestTimeout:     time.Hour,
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// AppConfig holds configuration for the App deployment controller.
type AppConfig struct {
	// Enabled controls whether the app deployment controller is started.
	Enabled bool `json:"enabled" pflag:",Enable app deployment controller"`

	// Namespace is the K8s namespace where KService CRDs are created.
	Namespace string `json:"namespace" pflag:",Namespace for app KServices"`

	// DefaultRequestTimeout is the request timeout applied to apps that don't specify one.
	DefaultRequestTimeout time.Duration `json:"defaultRequestTimeout" pflag:",Default request timeout for apps"`

	// MaxRequestTimeout is the hard cap on request timeout (Knative max is 3600s).
	MaxRequestTimeout time.Duration `json:"maxRequestTimeout" pflag:",Maximum allowed request timeout for apps"`
}

// Config holds the configuration for the Actions service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Kubernetes configuration
	Kubernetes KubernetesConfig `json:"kubernetes"`

	// WatchBufferSize is the buffer size for each worker's event channel.
	WatchBufferSize int `json:"watchBufferSize" pflag:",Buffer size for watch channels"`

	// WatchWorkers is the number of parallel event-processing goroutines.
	// Events for the same TaskAction are always routed to the same worker to preserve ordering.
	WatchWorkers int `json:"watchWorkers" pflag:",Number of parallel worker goroutines for processing watch events"`

	// RunServiceURL is the base URL for the internal run service.
	RunServiceURL string `json:"runServiceUrl" pflag:",Base URL of the internal run service"`

	// RecordFilterSize is the size of the bloom filter used to deduplicate RecordAction calls.
	RecordFilterSize int `json:"recordFilterSize" pflag:",Size of the oppo bloom filter for deduplicating RecordAction calls"`

	// Apps holds configuration for the app deployment controller.
	Apps AppConfig `json:"apps"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// KubernetesConfig holds Kubernetes client configuration
type KubernetesConfig struct {
	// Namespace where TaskAction CRs are located
	Namespace string `json:"namespace" pflag:",Kubernetes namespace for TaskAction CRs"`

	// KubeConfig path (optional - if empty, uses in-cluster config)
	KubeConfig string `json:"kubeconfig" pflag:",Path to kubeconfig file (optional)"`
}

// GetConfig returns the parsed actions configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
