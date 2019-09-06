package catalog

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
)

type Client interface {
	Get(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error)
	Put(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error
}

func NewCatalogClient(store storage.ProtobufStore) Client {
	catalogConfig := GetConfig()

	var catalogClient Client
	if catalogConfig.Type == LegacyDiscoveryType {
		catalogClient = NewLegacyDiscovery(catalogConfig.Endpoint, store)
	} else if catalogConfig.Type == NoOpDiscoveryType {
		catalogClient = NewNoOpDiscovery()
	}

	logger.Infof(context.Background(), "Created Catalog client, type: %v", catalogConfig.Type)
	return catalogClient
}
