package blob

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/any"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type ArtifactBlobStore struct {
	store *storage.DataStore
}

// OffloadArtifactCard stores the artifact card in the blob store
func (a *ArtifactBlobStore) OffloadArtifactCard(ctx context.Context, name, version string, card *any.Any) (storage.DataReference, error) {
	uri, err := a.store.ConstructReference(ctx, a.store.GetBaseContainerFQN(ctx), name, version)
	if err != nil {
		return "", fmt.Errorf("failed to construct data reference for [%s/%s] with err: %v", name, version, err)
	}
	err = a.store.WriteProtobuf(ctx, uri, storage.Options{}, card)
	if err != nil {
		return "", fmt.Errorf("failed to write protobuf to %s with err: %v", uri, err)
	}
	return uri, nil
}

func (a *ArtifactBlobStore) RetrieveArtifactCard(ctx context.Context, uri storage.DataReference) (*any.Any, error) {
	card := &any.Any{}
	err := a.store.ReadProtobuf(ctx, uri, card)
	if err != nil {
		return nil, fmt.Errorf("failed to read protobuf from %s with err: %v", uri, err)
	}
	return nil, nil
}

func NewArtifactBlobStore(ctx context.Context, scope promutils.Scope) ArtifactBlobStore {
	storageCfg := configuration.GetApplicationConfig().ArtifactBlobStoreConfig
	logger.Infof(ctx, "Initializing storage client with config [%+v]", storageCfg)

	dataStorageClient, err := storage.NewDataStore(&storageCfg, scope)
	if err != nil {
		logger.Error(ctx, "Failed to initialize storage config")
		panic(err)
	}
	return ArtifactBlobStore{
		store: dataStorageClient,
	}
}
