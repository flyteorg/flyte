package catalog

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/cacheservice"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

//go:generate pflags Config --default-var defaultConfig

const ConfigSectionKey = "catalog-cache"

var (
	defaultConfig = &Config{
		Type:               NoOpDiscoveryType,
		MaxRetries:         5,
		MaxPerRetryTimeout: config.Duration{Duration: 0},
		BackOffScalar:      100,
	}

	configSection = config.MustRegisterSection(ConfigSectionKey, defaultConfig)
)

type DiscoveryType = string

const (
	NoOpDiscoveryType DiscoveryType = "noop"
	DataCatalogType   DiscoveryType = "datacatalog"
	CacheServiceType  DiscoveryType = "cacheservice"
	FallbackType      DiscoveryType = "fallback"
)

type Config struct {
	Type               DiscoveryType   `json:"type" pflag:"\"noop\", Catalog Implementation to use"`
	Endpoint           string          `json:"endpoint" pflag:"\"\", Endpoint for catalog service"`
	CacheEndpoint      string          `json:"cache-endpoint" pflag:"\"\", Endpoint for cache service"`
	Insecure           bool            `json:"insecure" pflag:"false, Use insecure grpc connection"`
	MaxCacheAge        config.Duration `json:"max-cache-age" pflag:", Cache entries past this age will incur cache miss. 0 means cache never expires"`
	UseAdminAuth       bool            `json:"use-admin-auth" pflag:"false, Use the same gRPC credentials option as the flyteadmin client"`
	MaxRetries         int             `json:"max-retries" pflag:",Max number of gRPC retries"`
	MaxPerRetryTimeout config.Duration `json:"max-per-retry-timeout" pflag:",gRPC per retry timeout. O means no timeout."`
	BackOffScalar      int             `json:"backoff-scalar" pflag:",gRPC backoff scalar in milliseconds"`
	InlineCache        bool            `json:"inline-cache" pflag:"false, Attempt to use in-line cache"`

	// Set the gRPC service config formatted as a json string https://github.com/grpc/grpc/blob/master/doc/service_config.md
	// eg. {"loadBalancingConfig": [{"round_robin":{}}], "methodConfig": [{"name":[{"service": "foo", "method": "bar"}, {"service": "baz"}], "timeout": "1.000000001s"}]}
	// find the full schema here https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto#L625
	// Note that required packages may need to be preloaded to support certain service config. For example "google.golang.org/grpc/balancer/roundrobin" should be preloaded to have round-robin policy supported.
	DefaultServiceConfig string `json:"default-service-config" pflag:"\"\", Set the default service config for the catalog gRPC client"`
}

// GetConfig gets loaded config for Discovery
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func NewCacheClient(ctx context.Context, dataStore *storage.DataStore, authOpt ...grpc.DialOption) (catalog.Client, error) {
	catalogConfig := GetConfig()

	switch catalogConfig.Type {
	case CacheServiceType:
		return cacheservice.NewCacheClient(ctx, dataStore, catalogConfig.CacheEndpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.MaxRetries,
			catalogConfig.MaxPerRetryTimeout.Duration, catalogConfig.BackOffScalar, catalogConfig.InlineCache,
			catalogConfig.DefaultServiceConfig, authOpt...)
	case FallbackType:
		cacheClient, err := cacheservice.NewCacheClient(ctx, dataStore, catalogConfig.CacheEndpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.MaxRetries,
			catalogConfig.MaxPerRetryTimeout.Duration, catalogConfig.BackOffScalar, catalogConfig.InlineCache,
			catalogConfig.DefaultServiceConfig, authOpt...)
		if err != nil {
			return nil, err
		}
		catalogClient, err := datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			authOpt...)
		if err != nil {
			return nil, err
		}
		return cacheservice.NewFallbackClient(cacheClient, catalogClient)
	case DataCatalogType:
		return datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			authOpt...)
	case NoOpDiscoveryType, "":
		return NOOPCatalog{}, nil
	}
	return nil, fmt.Errorf("no such catalog type available: %s", catalogConfig.Type)
}
