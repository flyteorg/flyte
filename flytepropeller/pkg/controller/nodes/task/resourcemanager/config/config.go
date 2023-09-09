package config

import (
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
)

//go:generate pflags Config --default-var=defaultConfig

const configSectionKey = "resourcemanager"

type Type = string

const (
	TypeNoop  Type = "noop"
	TypeRedis Type = "redis"
)

var (
	defaultConfig = Config{
		Type: TypeNoop,
		// TODO: Noop Resource Manager doesn't use MaxQuota. Maybe we can remove it?
		ResourceMaxQuota: 1000,
	}

	configSection = config.MustRegisterSubSection(configSectionKey, &defaultConfig)
)

// Configs for Resource Manager
type Config struct {
	Type             Type        `json:"type" pflag:"noop,Which resource manager to use"`
	ResourceMaxQuota int         `json:"resourceMaxQuota" pflag:",Global limit for concurrent Qubole queries"`
	RedisConfig      RedisConfig `json:"redis" pflag:",Config for Redis resourcemanager."`
}

// Specific configs for Redis resource manager
// Ref: https://redis.io/topics/sentinel for information on how to fill in these fields.
type RedisConfig struct {
	HostPaths   []string `json:"hostPaths" pflag:",Redis hosts locations."`
	PrimaryName string   `json:"primaryName" pflag:",Redis primary name, fill in only if you are connecting to a redis sentinel cluster."`
	// deprecated: Please use HostPaths instead
	HostPath   string `json:"hostPath" pflag:",Redis host location"`
	HostKey    string `json:"hostKey" pflag:",Key for local Redis access"`
	MaxRetries int    `json:"maxRetries" pflag:",See Redis client options for more info"`
}

// Retrieves the current config value or default.
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func SetConfig(cfg *Config) error {
	return configSection.SetConfig(cfg)
}
