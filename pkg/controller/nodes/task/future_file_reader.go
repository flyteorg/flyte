package task

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
	"github.com/pkg/errors"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// TODO this file exists only until we need to support dynamic nodes instead of closure.
// Once closure migration is done, this file should be deleted.
const implicitFutureFileName = "futures.pb"
const implicitCompileWorkflowsName = "futures_compiled.pb"

type FutureFileReader struct {
	RemoteFileWorkflowStore
	loc      storage.DataReference
	cacheLoc storage.DataReference
	store    *storage.DataStore
}

func (f FutureFileReader) Exists(ctx context.Context) (bool, error) {
	metadata, err := f.store.Head(ctx, f.loc)
	// If no futures file produced, then declare success and return.
	if err != nil {
		logger.Warnf(ctx, "Failed to read futures file. Error: %v", err)
		return false, errors.Wrap(err, "Failed to do HEAD on futures file.")
	}
	return metadata.Exists(), nil
}

func (f FutureFileReader) Read(ctx context.Context) (*core.DynamicJobSpec, error) {
	djSpec := &core.DynamicJobSpec{}
	if err := f.store.ReadProtobuf(ctx, f.loc, djSpec); err != nil {
		logger.Warnf(ctx, "Failed to read futures file. Error: %v", err)
		return nil, errors.Wrap(err, "Failed to read futures protobuf file.")
	}

	return djSpec, nil
}

func (f FutureFileReader) CacheExists(ctx context.Context) (bool, error) {
	return f.RemoteFileWorkflowStore.Exists(ctx, f.cacheLoc)
}

func (f FutureFileReader) Cache(ctx context.Context, wf *v1alpha1.FlyteWorkflow) error {
	return f.RemoteFileWorkflowStore.Put(ctx, wf, f.cacheLoc)
}

func (f FutureFileReader) RetrieveCache(ctx context.Context) (*v1alpha1.FlyteWorkflow, error) {
	return f.RemoteFileWorkflowStore.Get(ctx, f.cacheLoc)
}

func NewRemoteFutureFileReader(ctx context.Context, dataDir storage.DataReference, store *storage.DataStore) (FutureFileReader, error) {
	loc, err := store.ConstructReference(ctx, dataDir, implicitFutureFileName)
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for futures file. Error: %v", err)
		return FutureFileReader{}, err
	}

	cacheLoc, err := store.ConstructReference(ctx, dataDir, implicitCompileWorkflowsName)
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for compile workflows file, error: %s", err)
		return FutureFileReader{}, err
	}
	return FutureFileReader{
		loc:                     loc,
		cacheLoc:                cacheLoc,
		store:                   store,
		RemoteFileWorkflowStore: NewRemoteWorkflowStore(store),
	}, nil
}
