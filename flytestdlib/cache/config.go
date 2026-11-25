package cache

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	redisconfig "github.com/flyteorg/flyte/v2/flytestdlib/redis"
)

//go:generate enumer --type=Type -json -yaml -trimprefix=Type
//go:generate pflags Config --default-var=defaultConfig

type Type uint8

const (
	TypeInMemoryFixedSize Type = iota
	TypeRedis
)

var (
	defaultConfig = &Config{
		Type: TypeInMemoryFixedSize,
		InMemoryFixedSize: InMemoryFixedSizeConfig{
			Size: resource.NewScaledQuantity(100, resource.Mega),
			DefaultExpiration: config.Duration{
				Duration: 24 * time.Hour,
			},
		},
		Redis: RedisConfig{
			DefaultExpiration: config.Duration{
				Duration: 24 * time.Hour,
			},
		},
	}

	configSection = config.MustRegisterSection("cache", defaultConfig)
)

type Config struct {
	// Type of cache to use
	Type Type `json:"type" pflag:",type, Type of cache to use"`

	// Config for in-memory cache
	InMemoryFixedSize InMemoryFixedSizeConfig `json:"inMemoryFixedSize" pflag:"-,Config for in-memory cache"`

	// Config for Redis cache
	Redis RedisConfig `json:"redis" pflag:"-,Config for Redis cache"`
}

// InMemoryFixedSizeConfig is a copy of ristretto.Config that can be used in config files (removed func references)
type InMemoryFixedSizeConfig struct {
	Size              *resource.Quantity `json:"size" pflag:"-,Cache size (in bytes) to allocate. Note the memory will be allocated immediately and will never grow or shrink (minimizing GCs)."`
	DefaultExpiration config.Duration    `json:"defaultExpiration" pflag:",Default expiration time for items"`
}

type RedisConfig struct {
	Options           RedisOptions    `json:"options" pflag:"-,Redis options."`
	DefaultExpiration config.Duration `json:"defaultExpiration" pflag:",Default expiration time for items."`
}

// RedisOptions is the shared Redis client configuration.
// Kept as an alias for callers using the cache package.
type RedisOptions = redisconfig.Config

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func MustRegisterSubsection(name string, cfg config.Config) config.Section {
	return configSection.MustRegisterSection(name, cfg)
}
