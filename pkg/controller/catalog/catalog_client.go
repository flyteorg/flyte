package catalog

import (
	"context"

	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/controller/catalog/datacatalog"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
)

type Client interface {
	Get(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error)
	Put(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error
}

func NewCatalogClient(ctx context.Context, store storage.ProtobufStore) (Client, error) {
	catalogConfig := GetConfig()

	var catalogClient Client
	var err error
	switch catalogConfig.Type {
	case LegacyDiscoveryType:
		catalogClient = NewLegacyDiscovery(catalogConfig.Endpoint, store)
	case DataCatalogType:
		catalogClient, err = datacatalog.NewDataCatalog(ctx, catalogConfig.Endpoint, catalogConfig.Secure, store)
		if err != nil {
			return nil, err
		}
	case NoOpDiscoveryType, "":
		catalogClient = NewNoOpDiscovery()
	default:
		return nil, fmt.Errorf("No such catalog type available: %v", catalogConfig.Type)
	}

	logger.Infof(context.Background(), "Created Catalog client, type: %v", catalogConfig.Type)
	return catalogClient, nil
}
