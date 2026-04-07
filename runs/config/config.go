package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
)

const configSectionKey = "runs"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8090,
		Host: "0.0.0.0",
	},
	WatchBufferSize:   100,
	ActionsServiceURL: "http://localhost:8090",
	StoragePrefix:     "file:///tmp/flyte/data",
	SeedProjects:      []string{"flytesnacks"},
	Domains: []DomainConfig{
		{ID: "development", Name: "Development"},
		{ID: "production", Name: "Production"},
		{ID: "staging", Name: "Staging"},
	},
	Apps: AppsConfig{
		InternalAppServiceURL: "http://localhost:8091",
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

// Config holds the configuration for the Runs service
type Config struct {
	// HTTP server configuration
	Server ServerConfig `json:"server"`

	// Database configuration
	Database database.DbConfig `json:"database"`

	// Watch/streaming settings
	WatchBufferSize int `json:"watchBufferSize" pflag:",Buffer size for watch streams"`

	// Actions service URL for enqueuing actions
	ActionsServiceURL string `json:"actionsServiceUrl" pflag:",URL of the actions service"`

	// StoragePrefix is the base URI for storing run data (inputs, outputs)
	// e.g. "s3://my-bucket" or "gs://my-bucket" or "file:///tmp/flyte/data"
	StoragePrefix string `json:"storagePrefix" pflag:",Base URI prefix for storing run inputs and outputs"`

	// SeedProjects are created at startup. If unset, the service seeds the default project list.
	SeedProjects []string `json:"seedProjects" pflag:",Projects to create by default at startup"`

	// Domains are injected into project responses (not stored per project row).
	Domains []DomainConfig `json:"domains"`

	// Apps holds configuration for the App service.
	Apps AppsConfig `json:"apps"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// DomainConfig defines a system domain to inject into project responses.
type DomainConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AppsConfig holds configuration for the App service in the runs (control plane).
type AppsConfig struct {
	// PublicURLPattern is a Go template for generating public ingress URLs.
	// Available variables: {{.Name}}, {{.Project}}, {{.Domain}}
	// Example: "https://{{.Name}}-{{.Project}}.apps.flyte.example.com"
	PublicURLPattern string `json:"publicUrlPattern" pflag:",URL pattern for app ingress"`

	// InternalAppServiceURL is the base URL of the InternalAppService (actions data plane).
	// In unified mode this is overridden by sc.BaseURL.
	InternalAppServiceURL string `json:"internalAppServiceUrl" pflag:",URL of the internal app service"`
}

// GetConfig returns the parsed runs configuration
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
