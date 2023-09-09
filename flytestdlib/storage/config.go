package storage

import (
	"context"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate pflags Config --default-var=defaultConfig

// Type defines the storage config type.
type Type = string

// The reserved config section key for storage.
const configSectionKey = "Storage"

const (
	TypeMemory Type = "mem"
	TypeS3     Type = "s3"
	TypeLocal  Type = "local"
	TypeMinio  Type = "minio"
	TypeStow   Type = "stow"
)

const (
	KiB int64 = 1024
	MiB int64 = 1024 * KiB
)

var (
	ConfigSection = config.MustRegisterSection(configSectionKey, defaultConfig)
	defaultConfig = &Config{
		Type: TypeS3,
		Limits: LimitsConfig{
			GetLimitMegabytes: 2,
		},
		Connection: ConnectionConfig{
			Region:   "us-east-1",
			AuthType: "iam",
		},
		MultiContainerEnabled: false,
	}
)

// Config is a common storage config.
type Config struct {
	Type Type `json:"type" pflag:",Sets the type of storage to configure [s3/minio/local/mem/stow]."`
	// Deprecated: Please use StowConfig instead
	Connection ConnectionConfig `json:"connection"`
	Stow       StowConfig       `json:"stow,omitempty" pflag:",Storage config for stow backend."`

	// Container here is misleading, it refers to a Bucket (AWS S3) like blobstore entity. In some terms it could be a table
	InitContainer string `json:"container" pflag:",Initial container (in s3 a bucket) to create -if it doesn't exist-.'"`
	// By default if this is not enabled, multiple containers are not supported by the storage layer. Only the configured `container` InitContainer will be allowed to requests data from. But, if enabled then data will be loaded to written to any
	// container specified in the DataReference.
	MultiContainerEnabled bool `json:"enable-multicontainer" pflag:",If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered"`
	// Caching is recommended to improve the performance of underlying systems. It caches the metadata and resolving
	// inputs is accelerated. The size of the cache is large so understand how to configure the cache.
	// TODO provide some default config choices
	// If this section is skipped, Caching is disabled
	Cache             CachingConfig    `json:"cache"`
	Limits            LimitsConfig     `json:"limits" pflag:",Sets limits for stores."`
	DefaultHTTPClient HTTPClientConfig `json:"defaultHttpClient" pflag:",Sets the default http client config."`
	SignedURL         SignedURLConfig  `json:"signedUrl" pflag:",Sets config for SignedURL."`
}

// SignedURLConfig encapsulates configs specifically used for SignedURL behavior.
type SignedURLConfig struct {
	StowConfigOverride map[string]string `json:"stowConfigOverride,omitempty" pflag:"-,Configuration for stow backend. Refer to github/flyteorg/stow"`
}

// HTTPClientConfig encapsulates common settings that can be applied to an HTTP Client.
type HTTPClientConfig struct {
	Headers map[string][]string `json:"headers" pflag:"-,Sets http headers to set on the http client."`
	Timeout config.Duration     `json:"timeout" pflag:",Sets time out on the http client."`
}

// ConnectionConfig defines connection configurations.
type ConnectionConfig struct {
	Endpoint   config.URL `json:"endpoint" pflag:",URL for storage client to connect to."`
	AuthType   string     `json:"auth-type" pflag:",Auth Type to use [iam,accesskey]."`
	AccessKey  string     `json:"access-key" pflag:",Access key to use. Only required when authtype is set to accesskey."`
	SecretKey  string     `json:"secret-key" pflag:",Secret to use when accesskey is set."`
	Region     string     `json:"region" pflag:",Region to connect to."`
	DisableSSL bool       `json:"disable-ssl" pflag:",Disables SSL connection. Should only be used for development."`
}

// StowConfig defines configs for stow as defined in github.com/flyteorg/stow
type StowConfig struct {
	Kind   string            `json:"kind,omitempty" pflag:",Kind of Stow backend to use. Refer to github/flyteorg/stow"`
	Config map[string]string `json:"config,omitempty" pflag:",Configuration for stow backend. Refer to github/flyteorg/stow"`
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

// LimitsConfig specifies limits for storage package.
type LimitsConfig struct {
	GetLimitMegabytes int64 `json:"maxDownloadMBs" pflag:",Maximum allowed download size (in MBs) per call."`
}

// GetConfig retrieve current global config for storage.
func GetConfig() *Config {
	if c, ok := ConfigSection.GetConfig().(*Config); ok {
		return c
	}

	logger.Warnf(context.TODO(), "Failed to retrieve config section [%v].", configSectionKey)
	return nil
}
