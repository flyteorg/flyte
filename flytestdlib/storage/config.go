package storage

import (
	"context"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/logger"
)

//go:generate pflags Config

// Defines the storage config type.
type Type = string

// The reserved config section key for storage.
const configSectionKey = "Storage"

const (
	TypeMemory Type = "mem"
	TypeS3     Type = "s3"
	TypeLocal  Type = "local"
	TypeMinio  Type = "minio"
)

const (
	KiB int64 = 1024
	MiB int64 = 1024 * KiB
)

var (
	ConfigSection = config.MustRegisterSection(configSectionKey, &Config{})
)

// A common storage config.
type Config struct {
	Type          Type             `json:"type" pflag:"\"s3\",Sets the type of storage to configure [s3/minio/local/mem]."`
	Connection    ConnectionConfig `json:"connection"`
	InitContainer string           `json:"container" pflag:",Initial container to create -if it doesn't exist-.'"`
	// Caching is recommended to improve the performance of underlying systems. It caches the metadata and resolving
	// inputs is accelerated. The size of the cache is large so understand how to configure the cache.
	// TODO provide some default config choices
	// If this section is skipped, Caching is disabled
	Cache  CachingConfig `json:"cache"`
	Limits LimitsConfig  `json:"limits" pflag:",Sets limits for stores."`
}

// Defines connection configurations.
type ConnectionConfig struct {
	Endpoint   config.URL `json:"endpoint" pflag:",URL for storage client to connect to."`
	AuthType   string     `json:"auth-type" pflag:"\"iam\",Auth Type to use [iam,accesskey]."`
	AccessKey  string     `json:"access-key" pflag:",Access key to use. Only required when authtype is set to accesskey."`
	SecretKey  string     `json:"secret-key" pflag:",Secret to use when accesskey is set."`
	Region     string     `json:"region" pflag:"\"us-east-1\",Region to connect to."`
	DisableSSL bool       `json:"disable-ssl" pflag:",Disables SSL connection. Should only be used for development."`
}

type CachingConfig struct {
	// Maximum size of the cache where the Blob store data is cached in-memory
	// Refer to https://github.com/coocood/freecache to understand how to set the value
	// If not specified or set to 0, cache is not used
	// NOTE: if Object sizes are larger than 1/1024 of the cache size, the entry will not be written to the cache
	// Also refer to https://github.com/coocood/freecache/issues/17 to understand how to set the cache
	MaxSizeMegabytes int `json:"max_size_mbs" pflag:",Maximum size of the cache where the Blob store data is cached in-memory. If not specified or set to 0, cache is not used"`
	// sets the garbage collection target percentage:
	// a collection is triggered when the ratio of freshly allocated data
	// to live data remaining after the previous collection reaches this percentage.
	// refer to https://golang.org/pkg/runtime/debug/#SetGCPercent
	// If not specified or set to 0, GC percent is not tweaked
	TargetGCPercent int `json:"target_gc_percent" pflag:",Sets the garbage collection target percentage."`
}

// Specifies limits for storage package.
type LimitsConfig struct {
	GetLimitMegabytes int64 `json:"maxDownloadMBs" pflag:"2,Maximum allowed download size (in MBs) per call."`
}

// Retrieve current global config for storage.
func GetConfig() *Config {
	if c, ok := ConfigSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(context.TODO(), "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
