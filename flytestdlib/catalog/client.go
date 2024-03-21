package catalog

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	pluginCatalog "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytestdlib/catalog/datacatalog"
)

func NewCatalogClient(ctx context.Context, authOpt ...grpc.DialOption) (pluginCatalog.Client, error) {
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
