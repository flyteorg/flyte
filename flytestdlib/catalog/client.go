package catalog

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	pluginCatalog "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flytestdlib/catalog/datacatalog"
)

func NewClient(ctx context.Context, authOpt ...grpc.DialOption) (pluginCatalog.Client, error) {
	catalogConfig := GetConfig()

	switch catalogConfig.Type {
	case TypeDataCatalog:
		return datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Insecure,
			catalogConfig.MaxCacheAge.Duration, catalogConfig.UseAdminAuth, catalogConfig.DefaultServiceConfig,
			authOpt...)
	case TypeNoOp, "":
		return NOOPCatalog{}, nil
	}

	return nil, fmt.Errorf("invalid catalog type %q", catalogConfig.Type)
}
