package config

import "github.com/flyteorg/flyte/v2/flytestdlib/config"

const configSectionKey = "cache_service"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8094,
		Host: "0.0.0.0",
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

type Config struct {
	Server ServerConfig `json:"server"`
}

type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
