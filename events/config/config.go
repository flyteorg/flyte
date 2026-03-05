package config

import "github.com/flyteorg/flyte/v2/flytestdlib/config"

const configSectionKey = "events"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8092,
		Host: "0.0.0.0",
	},
	QueueSize:     1000,
	WorkerCount:   4,
	RunServiceURL: "http://localhost:8090",
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds configuration for the Events service.
type Config struct {
	// HTTP server configuration.
	Server ServerConfig `json:"server"`

	// QueueSize is the number of events that can be buffered in memory.
	QueueSize int `json:"queueSize" pflag:",Buffer size for in-memory events queue"`

	// WorkerCount is the number of background workers forwarding events.
	WorkerCount int `json:"workerCount" pflag:",Number of background workers"`

	// RunServiceURL is the base URL for the internal run service.
	RunServiceURL string `json:"runServiceUrl" pflag:",Base URL of the internal run service"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed events configuration.
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
