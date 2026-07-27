package storage

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/stow/s3"
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
	TypeRedis  Type = "redis"
)

const (
	KiB int64 = 1024
	MiB int64 = 1024 * KiB
)

var (
	ConfigSection = config.MustRegisterSection(configSectionKey, defaultConfig)
	defaultConfig = &Config{
		Type: TypeStow,
		Stow: StowConfig{
			Kind: s3.Kind,
			Config: map[string]string{
				s3.ConfigRegion: "us-east-1",
			},
		},
		Limits: LimitsConfig{
			GetLimitMegabytes: 2,
		},
		DefaultHTTPClient: HTTPClientConfig{
			MaxIdleConns:        1024,
			MaxIdleConnsPerHost: 1024,
			IdleConnTimeout:     config.Duration{Duration: 90 * time.Second},
		},
		MultiContainerEnabled: false,
	}
)

// Config is a common storage config.
type Config struct {
	Type  Type        `json:"type" pflag:",Sets the type of storage [s3/minio/local/mem/stow/redis]."`
	Stow  StowConfig  `json:"stow,omitempty" pflag:",Storage config for stow backend."`
	Redis RedisConfig `json:"redis,omitempty" pflag:"-,Storage config for the redis backend."`

	// Container here is misleading, it refers to a Bucket (AWS S3) like blobstore entity. In some terms it could be a table
	InitContainer string `json:"container" pflag:",Initial container (in s3 a bucket) to create -if it doesn't exist-.'"`
	// By default if this is not enabled, multiple containers are not supported by the storage layer. Only the configured `container` InitContainer will be allowed to requests data from. But, if enabled then data will be loaded to written to any
	// container specified in the DataReference.
	MultiContainerEnabled bool `json:"enable-multicontainer" pflag:",If this is true, then the container argument is overlooked and redundant. This config will automatically open new connections to new containers/buckets as they are encountered"`
	// Schemes provides optional per-scheme backend overrides for the multi-scheme DataStore. It is
	// keyed by URL scheme (e.g. "s3", "gs", "abfs", "redis"). The DataStore lazily instantiates a
	// backend the first time a reference with a given scheme is seen and memoizes it for reuse, so a
	// single DataStore can serve any scheme/container thrown at it. A scheme absent from this map is
	// dialed with ambient credentials (the provider's default credential chain — IAM/instance profile,
	// GCP ADC, workload identity, env vars). The primary scheme (derived from Type) continues to be
	// configured by the top-level Connection/Stow/Redis fields and owns GetBaseContainerFQN.
	Schemes map[string]SchemeConfig `json:"schemes,omitempty" pflag:"-,Optional per-scheme backend overrides keyed by URL scheme."`
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
	// Zero values for the transport settings below fall back to the http.DefaultTransport values.
	MaxIdleConns        int             `json:"maxIdleConns" pflag:",Maximum number of idle connections across all hosts. Zero means use the http.DefaultTransport value."`
	MaxIdleConnsPerHost int             `json:"maxIdleConnsPerHost" pflag:",Maximum number of idle connections per host. Zero means use the http.DefaultTransport value."`
	MaxConnsPerHost     int             `json:"maxConnsPerHost" pflag:",Maximum number of connections per host; new requests block at the limit. Zero means no limit."`
	IdleConnTimeout     config.Duration `json:"idleConnTimeout" pflag:",Maximum amount of time an idle connection remains open. Zero means use the http.DefaultTransport value."`
}

// RedisConfig defines the connection for a redis-backed raw store, selected with type: redis.
// Objects are stored as redis string values keyed by the path portion of the DataReference
// (redis://<addr>/<key>), which keeps references interoperable with the flyte-sdk redis plugin.
type RedisConfig struct {
	// Addr is the host:port of the redis server.
	Addr string `json:"addr"`
	// Username for redis ACL authentication (redis 6+), if required.
	Username string `json:"username"`
	// Password for redis authentication, if required.
	Password string `json:"password"`
	// DB is the logical database to select after connecting.
	DB int `json:"db"`
}

// StowConfig defines configs for stow as defined in github.com/flyteorg/stow
type StowConfig struct {
	Kind   string            `json:"kind,omitempty" pflag:",Kind of Stow backend to use. Refer to github/flyteorg/stow"`
	Config map[string]string `json:"config,omitempty" pflag:",Configuration for stow backend. Refer to github/flyteorg/stow"`
}

// SchemeConfig overrides how a single URL scheme is dialed by the multi-scheme DataStore. All fields
// are optional: when omitted, the scheme is dialed with ambient credentials and a stow kind derived
// from the scheme. Used for backends that need explicit credentials/endpoints (e.g. minio, a redis
// address that differs from the reference host, or a non-default account).
type SchemeConfig struct {
	// Kind overrides the stow kind for this scheme. Defaults to the kind derived from the scheme
	// (s3->s3, gs->google, abfs/abfss->azure, os->oracle, sw->swift, file->local). Ignored for redis.
	Kind string `json:"kind,omitempty"`
	// Config is the stow ConfigMap (credentials, region, endpoint, ...) passed to stow.Dial for this
	// scheme. When empty, the scheme is dialed with ambient credentials.
	Config map[string]string `json:"config,omitempty"`
	// Redis configures the connection when the scheme is redis. When nil, the redis address is taken
	// from the top-level Redis.Addr, falling back to the host portion of the reference.
	Redis *RedisConfig `json:"redis,omitempty"`
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
