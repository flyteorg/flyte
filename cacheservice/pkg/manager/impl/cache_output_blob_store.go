package impl

import (
	"context"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var (
	_ interfaces.CacheOutputBlobStore = &cacheOutputBlobStore{}
)

type cacheOutputBlobStore struct {
	store         *storage.DataStore
	storagePrefix storage.DataReference
}

func (m *cacheOutputBlobStore) Create(ctx context.Context, key string, output *core.LiteralMap) (string, error) {
	reference, err := m.store.ConstructReference(ctx, m.storagePrefix, key)
	if err != nil {
		return "", errors.NewCacheServiceErrorf(codes.Internal, "Unable to construct reference for key %s, err %v", key, err)
	}

	err = m.store.WriteProtobuf(ctx, reference, storage.Options{}, output)
	if err != nil {
		return "", errors.NewCacheServiceErrorf(codes.Internal, "Unable to store cache data in location %s, err %v", reference.String(), err)
	}
	return reference.String(), nil
}

func (m *cacheOutputBlobStore) Delete(ctx context.Context, uri string) error {
	reference, err := m.store.ConstructReference(ctx, m.storagePrefix, uri)
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Unable to construct reference for uri %s, err %v", uri, err)
	}
	err = m.store.Delete(ctx, reference)
	if err != nil {
		return errors.NewCacheServiceErrorf(codes.Internal, "Unable to store delete data in location %s, err %v", reference.String(), err)
	}
	return nil
}

func NewCacheOutputStore(store *storage.DataStore, storagePrefix storage.DataReference) interfaces.CacheOutputBlobStore {
	return &cacheOutputBlobStore{
		store:         store,
		storagePrefix: storagePrefix,
	}
}
