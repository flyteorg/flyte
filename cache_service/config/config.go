package config

import (
	"time"

	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const configSectionKey = "cache_service"

//go:generate pflags Config --default-var=defaultConfig

var defaultConfig = &Config{
	Server: ServerConfig{
		Port: 8094,
		Host: "0.0.0.0",
	},
	HeartbeatGracePeriodMultiplier: 3,
	MaxReservationHeartbeat:        stdconfig.Duration{Duration: 10 * time.Second},
}

var configSection = stdconfig.MustRegisterSection(configSectionKey, defaultConfig)

type Config struct {
	Server ServerConfig `json:"server"`

	HeartbeatGracePeriodMultiplier int                `json:"heartbeatGracePeriodMultiplier" pflag:",Number of heartbeats before a reservation expires without an extension."`
	MaxReservationHeartbeat        stdconfig.Duration `json:"maxReservationHeartbeat" pflag:",Maximum reservation heartbeat interval."`
}

type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
