package catalog

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flytestdlib/config"
	"google.golang.org/grpc"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/catalog/datacatalog"
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
}

// Gets loaded config for Discovery
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func NewCatalogClient(ctx context.Context, authOpt grpc.DialOption) (catalog.Client, error) {
	catalogConfig := GetConfig()

	switch catalogConfig.Type {
	case DataCatalogType:
		return datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure, catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, authOpt)
	case NoOpDiscoveryType, "":
		return NOOPCatalog{}, nil
	}
	return nil, fmt.Errorf("no such catalog type available: %s", catalogConfig.Type)
}
