package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
)

const configSectionKey = "events"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8092,
		Host: "0.0.0.0",
	},
	RunService: serviceclient.ServiceConfig{URL: "http://localhost:8090"},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds configuration for the Events service.
type Config struct {
	// HTTP server configuration.
	Server ServerConfig `json:"server"`

	// RunService configures the internal run service client.
	RunService serviceclient.ServiceConfig `json:"runService" pflag:",Internal run service client configuration"`
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
