package config

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/database"
)

const configSectionKey = "queue"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8089,
		Host: "0.0.0.0",
	},
	MaxQueueSize: 10000,
	WorkerCount:  10,
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Queue service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Database configuration (reuses flytestdlib)
	Database database.DbConfig `json:"database"`

	// Queue specific settings
	MaxQueueSize int `json:"maxQueueSize" pflag:",Maximum number of queued actions"`
	WorkerCount  int `json:"workerCount" pflag:",Number of worker goroutines for processing queue"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed queue configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
