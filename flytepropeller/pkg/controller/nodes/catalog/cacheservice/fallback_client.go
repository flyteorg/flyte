package cacheservice

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	catalogIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	_ catalog.Client = &FallbackClient{}
)

type FallbackClient struct {
	cacheClient   catalog.Client
	catalogClient catalog.Client
	dataStore     *storage.DataStore
}

func (c *FallbackClient) GetReservationCache(ownerID string) catalog.ReservationCache {
	return c.cacheClient.GetReservationCache(ownerID)
}

func (c *FallbackClient) UpdateReservationCache(ownerID string, entry catalog.ReservationCache) {
	c.cacheClient.UpdateReservationCache(ownerID, entry)
}

func (c *FallbackClient) GetOrExtendReservation(ctx context.Context, key catalog.Key, ownerID string, heartbeatInterval time.Duration) (*catalogIdl.Reservation, error) {
	return c.cacheClient.GetOrExtendReservation(ctx, key, ownerID, heartbeatInterval)
}

func (c *FallbackClient) ReleaseReservation(ctx context.Context, key catalog.Key, ownerID string) error {
	return c.cacheClient.ReleaseReservation(ctx, key, ownerID)
}

func (c *FallbackClient) generateFileOutputReader(ctx context.Context, key catalog.Key, catalogEntry catalog.Entry) (ioutils.RemoteFileOutputReader, error) {
	cacheKey, err := GenerateCacheKey(ctx, key)
	if err != nil {
		return ioutils.RemoteFileOutputReader{}, err
	}

	reference, err := c.dataStore.ConstructReference(ctx, c.dataStore.GetBaseContainerFQN(ctx), "cached_outputs", cacheKey)
	if err != nil {
		return ioutils.RemoteFileOutputReader{}, err
	}

	outputPath := ioutils.NewReadOnlyOutputFilePaths(ctx, c.dataStore, reference)
	outputs, ee, err := catalogEntry.GetOutputs().Read(ctx)
	if err != nil || ee != nil {
		return ioutils.RemoteFileOutputReader{}, err
	}

	err = c.dataStore.WriteProtobuf(ctx, outputPath.GetOutputPath(), storage.Options{}, outputs)
	if err != nil {
		return ioutils.RemoteFileOutputReader{}, err
	}

	outputReader := ioutils.NewRemoteFileOutputReader(ctx, c.dataStore, outputPath, int64(999999999))
	return outputReader, nil
}

func (c *FallbackClient) Get(ctx context.Context, key catalog.Key) (catalog.Entry, error) {
	cacheEntry, err := c.cacheClient.Get(ctx, key)
	if err != nil {
		if grpcStatus, ok := status.FromError(err); ok && grpcStatus.Code() == codes.NotFound {
			logger.Debugf(ctx, "Cache miss for key [%s], falling back to catalog", key)
		} else {
			return catalog.Entry{}, err
		}
	} else {
		return cacheEntry, err
	}

	catalogEntry, err := c.catalogClient.Get(ctx, key)
	if err != nil {
		return catalog.Entry{}, err
	}

	metadata := catalog.Metadata{
		CreatedAt: catalogEntry.GetStatus().GetMetadata().GetCreatedAt(),
	}
	identifier := catalogEntry.GetStatus().GetMetadata().GetSourceExecution()
	// TODO - pvditt update this when sub-workflow caching is supported
	switch identifier.(type) {
	case *core.CatalogMetadata_SourceTaskExecution:
		metadata.TaskExecutionIdentifier = catalogEntry.GetStatus().GetMetadata().GetSourceTaskExecution()
	}

	outputReader, err := c.generateFileOutputReader(ctx, key, catalogEntry)
	if err != nil {
		logger.Warnf(ctx, "Failed to generate output reader for key [%s] when falling back, err: %v", key, err)
		return catalogEntry, nil
	}

	_, err = c.cacheClient.Put(ctx, key, outputReader, metadata)
	if err != nil {
		logger.Warnf(ctx, "Failed to cache entry for key [%s] when falling back, err: %v", key, err)
	}

	return catalogEntry, nil
}

func (c *FallbackClient) Update(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return c.cacheClient.Update(ctx, key, reader, metadata)
}

func (c *FallbackClient) Put(ctx context.Context, key catalog.Key, reader io.OutputReader, metadata catalog.Metadata) (catalog.Status, error) {
	return c.cacheClient.Put(ctx, key, reader, metadata)
}

func NewFallbackClient(cacheClient catalog.Client, catalogClient catalog.Client, dataStore *storage.DataStore) (*FallbackClient, error) {
	return &FallbackClient{
		cacheClient:   cacheClient,
		catalogClient: catalogClient,
		dataStore:     dataStore,
	}, nil
}
