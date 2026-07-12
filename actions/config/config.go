package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "actions"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8091,
		Host: "0.0.0.0",
	},
	Kubernetes: app.K8sConfig{
		Namespace: "flyte",
		QPS:       1000,
		Burst:     2000,
		Timeout:   "30s",
	},
	// NOTE: total buffering is WatchWorkers * WatchBufferSize watch.Event entries.
	// Each entry includes a DeepCopy() of the TaskAction; tune carefully to avoid large memory spikes under backlog.
	WatchBufferSize: 1000,
	WatchWorkers:    100,
	RunServiceURL:   "http://localhost:8090",
	// LRU of recorded action keys (~1M live actions, ~100 MiB) to dedup RecordAction.
	// Recency eviction keeps the live working set, so it need not hold all cumulative actions.
	RecordFilterSize: 1 << 20,
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Actions service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Kubernetes configuration (namespace where TaskAction CRs live; the unified
	// binary shares the manager's client, this is read by the standalone cmd).
	Kubernetes app.K8sConfig `json:"kubernetes"`

	// WatchBufferSize is the buffer size for each worker's event channel.
	WatchBufferSize int `json:"watchBufferSize" pflag:",Buffer size for watch channels"`

	// WatchWorkers is the number of parallel event-processing goroutines.
	// Events for the same TaskAction are always routed to the same worker to preserve ordering.
	WatchWorkers int `json:"watchWorkers" pflag:",Number of parallel worker goroutines for processing watch events"`

	// RunServiceURL is the base URL for the internal run service.
	RunServiceURL string `json:"runServiceUrl" pflag:",Base URL of the internal run service"`

	// RecordFilterSize is the size of the bloom filter used to deduplicate RecordAction calls.
	RecordFilterSize int `json:"recordFilterSize" pflag:",Size of the oppo bloom filter for deduplicating RecordAction calls"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed actions configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
