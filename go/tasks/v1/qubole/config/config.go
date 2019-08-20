package config

//go:generate pflags Config

import (
	"time"

	"github.com/lyft/flytestdlib/config"

	pluginsConfig "github.com/lyft/flyteplugins/go/tasks/v1/config"
)

const quboleConfigSectionKey = "qubole"

var (
	defaultConfig = Config{
		QuboleLimit:            200,
		LruCacheSize:           1000,
		LookasideBufferPrefix:  "ql",
		LookasideExpirySeconds: config.Duration{Duration: time.Hour * 24},
	}

	quboleConfigSection = pluginsConfig.MustRegisterSubSection(quboleConfigSectionKey, &defaultConfig)
)

// Qubole plugin configs
type Config struct {
	QuboleTokenPath        string          `json:"quboleTokenPath" pflag:",Where to find the Qubole secret"`
	ResourceManagerType    string          `json:"resourceManagerType" pflag:"noop,Which resource manager to use"`
	RedisHostPath          string          `json:"redisHostPath" pflag:",Redis host location"`
	RedisHostKey           string          `json:"redisHostKey" pflag:",Key for local Redis access"`
	RedisMaxRetries        int             `json:"redisMaxRetries" pflag:",See Redis client options for more info"`
	QuboleLimit            int             `json:"quboleLimit" pflag:",Global limit for concurrent Qubole queries"`
	LruCacheSize           int             `json:"lruCacheSize" pflag:",Size of the AutoRefreshCache"`
	LookasideBufferPrefix  string          `json:"lookasideBufferPrefix" pflag:",Prefix used for lookaside buffer"`
	LookasideExpirySeconds config.Duration `json:"lookasideExpirySeconds" pflag:",TTL for lookaside buffer if supported"`
}

// Retrieves the current config value or default.
func GetQuboleConfig() *Config {
	return quboleConfigSection.GetConfig().(*Config)
}

func SetQuboleConfig(cfg *Config) error {
	return quboleConfigSection.SetConfig(cfg)
}
