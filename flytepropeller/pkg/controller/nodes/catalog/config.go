package catalog

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flytestdlib/config"
	"google.golang.org/grpc"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/catalog/datacatalog"
)

//go:generate pflags Config --default-var defaultConfig

const ConfigSectionKey = "catalog-cache"

var (
	defaultConfig = &Config{
		Type: NoOpDiscoveryType,
	}

	configSection = config.MustRegisterSection(ConfigSectionKey, defaultConfig)
)

type DiscoveryType = string

const (
	NoOpDiscoveryType DiscoveryType = "noop"
	DataCatalogType   DiscoveryType = "datacatalog"
)

type Config struct {
	Type         DiscoveryType   `json:"type" pflag:"\"noop\", Catalog Implementation to use"`
	Endpoint     string          `json:"endpoint" pflag:"\"\", Endpoint for catalog service"`
	Insecure     bool            `json:"insecure" pflag:"false, Use insecure grpc connection"`
	MaxCacheAge  config.Duration `json:"max-cache-age" pflag:", Cache entries past this age will incur cache miss. 0 means cache never expires"`
	UseAdminAuth bool            `json:"use-admin-auth" pflag:"false, Use the same gRPC credentials option as the flyteadmin client"`

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

func NewCatalogClient(ctx context.Context, authOpt ...grpc.DialOption) (catalog.Client, error) {
	catalogConfig := GetConfig()

	switch catalogConfig.Type {
	case DataCatalogType:
		return datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			authOpt...)
	case NoOpDiscoveryType, "":
		return NOOPCatalog{}, nil
	}
	return nil, fmt.Errorf("no such catalog type available: %s", catalogConfig.Type)
}
