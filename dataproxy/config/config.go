package config

import (
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"k8s.io/apimachinery/pkg/api/resource"
)

const configSectionKey = "dataproxy"

//go:generate pflags DataProxyConfig --default-var=defaultConfig

var defaultConfig = &DataProxyConfig{
	Server: ServerConfig{
		Port: 8088,
		Host: "0.0.0.0",
	},
	Upload: DataProxyUploadConfig{
		MaxSize:               resource.MustParse("100Mi"),
		MaxExpiresIn:          config.Duration{Duration: 3600000000000}, // 1 hour
		DefaultFileNameLength: 20,
		StoragePrefix:         "uploads",
	},
	Download: DataProxyDownloadConfig{
		MaxExpiresIn: config.Duration{Duration: 3600000000000}, // 1 hour
	},
	RunService: serviceclient.ServiceConfig{URL: "http://localhost:8090"},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

type DataProxyConfig struct {
	// Server configures the standalone DataProxy service HTTP listener.
	Server ServerConfig `json:"server"`

	Upload   DataProxyUploadConfig   `json:"upload" pflag:",Defines data proxy upload configuration."`
	Download DataProxyDownloadConfig `json:"download" pflag:",Defines data proxy download configuration."`

	// RunService configures the Runs service client. The Runs component also
	// hosts the Task, Trigger, and Project APIs used by dataproxy.
	RunService serviceclient.ServiceConfig `json:"runService" pflag:",Runs service client configuration."`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed data proxy configuration
func GetConfig() *DataProxyConfig {
	return configSection.GetConfig().(*DataProxyConfig)
}

type DataProxyDownloadConfig struct {
	MaxExpiresIn config.Duration `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
}

type DataProxyUploadConfig struct {
	MaxSize               resource.Quantity `json:"maxSize" pflag:",Maximum allowed upload size."`
	MaxExpiresIn          config.Duration   `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
	DefaultFileNameLength int               `json:"defaultFileNameLength" pflag:",Default length for the generated file name if file name not provided in the request."`
	StoragePrefix         string            `json:"storagePrefix" pflag:",Storage prefix to use for all upload requests."`
}
