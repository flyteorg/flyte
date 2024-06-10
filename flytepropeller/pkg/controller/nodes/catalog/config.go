package catalog

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

//go:generate pflags Config --default-var defaultConfig

const ConfigSectionKey = "catalog-cache"

var (
	defaultConfig = Config{
		Type:          NoOpDiscoveryType,
		MaxRetries:    5,
		BackoffScalar: 100,
		BackoffJitter: "0.1",
	}

	configSection = config.MustRegisterSectionWithUpdates(ConfigSectionKey, &defaultConfig, func(ctx context.Context, newValue config.Config) {
		if newValue.(*Config).MaxRetries < 0 {
			logger.Panicf(ctx, "Admin configuration given with negative gRPC retry value.")
		}

		if jitter, err := strconv.ParseFloat(newValue.(*Config).BackoffJitter, 64); err != nil || jitter < 0 || jitter > 1 {
			logger.Panicf(ctx, "Invalid jitter value [%v]. Must be between 0 and 1.", jitter)
		}
	})
)

type DiscoveryType = string

const (
	NoOpDiscoveryType DiscoveryType = "noop"
	DataCatalogType   DiscoveryType = "datacatalog"
)

type Config struct {
	Type          DiscoveryType   `json:"type" pflag:"\"noop\", Catalog Implementation to use"`
	Endpoint      string          `json:"endpoint" pflag:"\"\", Endpoint for catalog service"`
	Insecure      bool            `json:"insecure" pflag:"false, Use insecure grpc connection"`
	MaxCacheAge   config.Duration `json:"max-cache-age" pflag:", Cache entries past this age will incur cache miss. 0 means cache never expires"`
	UseAdminAuth  bool            `json:"use-admin-auth" pflag:"false, Use the same gRPC credentials option as the flyteadmin client"`
	MaxRetries    int             `json:"max-retries" pflag:",The max number of retries for event recording."`
	BackoffScalar int             `json:"base-scalar" pflag:",The base/scalar backoff duration in milliseconds for event recording retries."`
	BackoffJitter string          `json:"backoff-jitter" pflag:",A string representation of a floating point number between 0 and 1 specifying the jitter factor for event recording retries."`

	// Set the gRPC service config formatted as a json string https://github.com/grpc/grpc/blob/master/doc/service_config.md
	// eg. {"loadBalancingConfig": [{"round_robin":{}}], "methodConfig": [{"name":[{"service": "foo", "method": "bar"}, {"service": "baz"}], "timeout": "1.000000001s"}]}
	// find the full schema here https://github.com/grpc/grpc-proto/blob/master/grpc/service_config/service_config.proto#L625
	// Note that required packages may need to be preloaded to support certain service config. For example "google.golang.org/grpc/balancer/roundrobin" should be preloaded to have round-robin policy supported.
	DefaultServiceConfig string `json:"default-service-config" pflag:"\"\", Set the default service config for the catalog gRPC client"`
}

func (c Config) GetBackoffJitter(ctx context.Context) float64 {
	jitter, err := strconv.ParseFloat(c.BackoffJitter, 64)
	if err != nil {
		logger.Warnf(ctx, "Failed to parse backoff jitter [%v]. Error: %v", c.BackoffJitter, err)
		return 0.1
	}

	return jitter
}

// GetConfig gets loaded config for Discovery
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func NewCatalogClient(ctx context.Context, authOpt ...grpc.DialOption) (catalog.Client, error) {
	catalogConfig := GetConfig()

	switch catalogConfig.Type {
	case DataCatalogType:
		return datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			uint(catalogConfig.MaxRetries), catalogConfig.BackoffScalar, catalogConfig.GetBackoffJitter(ctx), authOpt...)
	case NoOpDiscoveryType, "":
		return NOOPCatalog{}, nil
	}
	return nil, fmt.Errorf("no such catalog type available: %s", catalogConfig.Type)
}
