package catalog

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	catalog2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/cacheservice"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func NewCacheClient(ctx context.Context, dataStore *storage.DataStore, authOpt ...grpc.DialOption) (catalog2.Client, error) {
	catalogConfig := config.GetConfig()

	switch catalogConfig.Type {
	case config.CacheServiceType:
		return cacheservice.NewCacheClient(ctx, dataStore, catalogConfig, false, authOpt...)
	case config.CacheServiceV2Type:
		return cacheservice.NewCacheClient(ctx, dataStore, catalogConfig, true, authOpt...)
	case config.FallbackType:
		cacheClient, err := cacheservice.NewCacheClient(ctx, dataStore, catalogConfig, false, authOpt...)
		if err != nil {
			return nil, err
		}
		catalogClient, err := datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			uint(catalogConfig.MaxRetries), catalogConfig.BackoffScalar, catalogConfig.GetBackoffJitter(ctx),
			catalogConfig.ReservationMaxCacheSize, authOpt...)
		if err != nil {
			return nil, err
		}
		return cacheservice.NewFallbackClient(cacheClient, catalogClient, dataStore)
	case config.DataCatalogType:
		return datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			uint(catalogConfig.MaxRetries), catalogConfig.BackoffScalar, catalogConfig.GetBackoffJitter(ctx),
			catalogConfig.ReservationMaxCacheSize, authOpt...)
	case config.NoOpDiscoveryType, "":
		return NOOPCatalog{}, nil
	}
	return nil, fmt.Errorf("no such catalog type available: %s", catalogConfig.Type)
}
