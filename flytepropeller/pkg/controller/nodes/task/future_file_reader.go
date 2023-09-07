package task

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/utils"
)

// TODO this file exists only until we need to support dynamic nodes instead of closure.
// Once closure migration is done, this file should be deleted.
const implicitFutureFileName = "futures.pb"
const implicitCompileWorkflowsName = "futures_compiled.pb"
const implicitCompiledWorkflowClosureName = "dynamic_compiled.pb"

type FutureFileReader struct {
	RemoteFileWorkflowStore
	loc                    storage.DataReference
	flyteWfCRDCacheLoc     storage.DataReference
	flyteWfClosureCacheLoc storage.DataReference
	store                  *storage.DataStore
}

func (f FutureFileReader) GetLoc() storage.DataReference {
	return f.loc
}

func (f FutureFileReader) Exists(ctx context.Context) (bool, error) {
	metadata, err := f.store.Head(ctx, f.loc)
	// If no futures file produced, then declare success and return.
	if err != nil {
		logger.Warnf(ctx, "Failed to read futures file. Error: %v", err)
		return false, errors.Wrapf(utils.ErrorCodeUser, err, "Failed to do HEAD on futures file.")
	}
	return metadata.Exists(), nil
}

func (f FutureFileReader) Read(ctx context.Context) (*core.DynamicJobSpec, error) {
	djSpec := &core.DynamicJobSpec{}
	if err := f.store.ReadProtobuf(ctx, f.loc, djSpec); err != nil {
		logger.Warnf(ctx, "Failed to read futures file. Error: %v", err)
		return nil, errors.Wrapf(utils.ErrorCodeSystem, err, "Failed to read futures protobuf file.")
	}

	return djSpec, nil
}

func (f FutureFileReader) CacheExists(ctx context.Context) (bool, error) {
	exists, err := f.RemoteFileWorkflowStore.Exists(ctx, f.flyteWfCRDCacheLoc)
	if err != nil || !exists {
		return exists, err
	}
	return f.RemoteFileWorkflowStore.Exists(ctx, f.flyteWfClosureCacheLoc)
}

func (f FutureFileReader) Cache(ctx context.Context, wf *v1alpha1.FlyteWorkflow, workflowClosure *core.CompiledWorkflowClosure) error {
	err := f.RemoteFileWorkflowStore.PutFlyteWorkflowCRD(ctx, wf, f.flyteWfCRDCacheLoc)
	if err != nil {
		return err
	}
	return f.RemoteFileWorkflowStore.PutCompiledFlyteWorkflow(ctx, workflowClosure, f.flyteWfClosureCacheLoc)
}

type CacheContents struct {
	WorkflowCRD      *v1alpha1.FlyteWorkflow
	CompiledWorkflow *core.CompiledWorkflowClosure
}

func (f FutureFileReader) RetrieveCache(ctx context.Context) (CacheContents, error) {
	workflowCRD, err := f.RemoteFileWorkflowStore.GetWorkflowCRD(ctx, f.flyteWfCRDCacheLoc)
	if err != nil {
		return CacheContents{}, err
	}
	compiledWorkflow, err := f.RemoteFileWorkflowStore.GetCompiledWorkflow(ctx, f.flyteWfClosureCacheLoc)
	if err != nil {
		return CacheContents{}, err
	}
	return CacheContents{
		WorkflowCRD:      workflowCRD,
		CompiledWorkflow: compiledWorkflow,
	}, nil
}

func NewRemoteFutureFileReader(ctx context.Context, dataDir storage.DataReference, store *storage.DataStore) (FutureFileReader, error) {
	loc, err := store.ConstructReference(ctx, dataDir, implicitFutureFileName)
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for futures file. Error: %v", err)
		return FutureFileReader{}, errors.Wrapf(utils.ErrorCodeSystem, err, "failed to construct data path")
	}

	flyteWfCRDCacheLoc, err := store.ConstructReference(ctx, dataDir, implicitCompileWorkflowsName)
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for compile workflows file, error: %s", err)
		return FutureFileReader{}, errors.Wrapf(utils.ErrorCodeSystem, err, "failed to construct reference for workflow CRD cache location")
	}
	flyteWfClosureCacheLoc, err := store.ConstructReference(ctx, dataDir, implicitCompiledWorkflowClosureName)
	if err != nil {
		logger.Warnf(ctx, "Failed to construct data path for compile workflows file, error: %s", err)
		return FutureFileReader{}, errors.Wrapf(utils.ErrorCodeSystem, err, "failed to construct reference for compiled workflow closure cache location")
	}

	return FutureFileReader{
		loc:                     loc,
		flyteWfCRDCacheLoc:      flyteWfCRDCacheLoc,
		flyteWfClosureCacheLoc:  flyteWfClosureCacheLoc,
		store:                   store,
		RemoteFileWorkflowStore: NewRemoteWorkflowStore(store),
	}, nil
}
