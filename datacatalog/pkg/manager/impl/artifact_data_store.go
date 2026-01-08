package impl

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const artifactDataFile = "data.pb"
const futureDataFile = "future.pb"
const futureDataName = "future"

// ArtifactDataStore stores and retrieves ArtifactData values in a data.pb
type ArtifactDataStore interface {
	PutData(ctx context.Context, artifact *datacatalog.Artifact, data *datacatalog.ArtifactData) (storage.DataReference, error)
	PutFutureData(ctx context.Context, artifact *datacatalog.Artifact, djspec *core.DynamicJobSpec) (storage.DataReference, error)
	GetFutureData(ctx context.Context, dataModel models.ArtifactData) (*core.DynamicJobSpec, error)
	GetData(ctx context.Context, dataModel models.ArtifactData) (*core.Literal, error)
	DeleteData(ctx context.Context, dataModel models.ArtifactData) error
}

type artifactDataStore struct {
	store         *storage.DataStore
	storagePrefix storage.DataReference
}

func (m *artifactDataStore) getDataLocation(ctx context.Context, artifact *datacatalog.Artifact, data *datacatalog.ArtifactData) (storage.DataReference, error) {
	dataset := artifact.GetDataset()
	return m.store.ConstructReference(ctx, m.storagePrefix, dataset.GetProject(), dataset.GetDomain(), dataset.GetName(), dataset.GetVersion(), artifact.GetId(), data.GetName(), artifactDataFile)
}

func (m *artifactDataStore) getFutureDataLocation(ctx context.Context, artifact *datacatalog.Artifact) (storage.DataReference, error) {
	dataset := artifact.GetDataset()
	return m.store.ConstructReference(ctx, m.storagePrefix, dataset.GetProject(), dataset.GetDomain(), dataset.GetName(), dataset.GetVersion(), artifact.GetId(), futureDataName, futureDataFile)
}

// Store marshalled data in data.pb under the storage prefix
func (m *artifactDataStore) PutData(ctx context.Context, artifact *datacatalog.Artifact, data *datacatalog.ArtifactData) (storage.DataReference, error) {
	dataLocation, err := m.getDataLocation(ctx, artifact, data)
	if err != nil {
		return "", errors.NewDataCatalogErrorf(codes.Internal, "Unable to generate data location %s, err %v", dataLocation.String(), err)
	}
	err = m.store.WriteProtobuf(ctx, dataLocation, storage.Options{}, data.GetValue())
	if err != nil {
		return "", errors.NewDataCatalogErrorf(codes.Internal, "Unable to store artifact data in location %s, err %v", dataLocation.String(), err)
	}

	return dataLocation, nil
}

// Store marshalled future data in future.pb under the future storage prefix
func (m *artifactDataStore) PutFutureData(ctx context.Context, artifact *datacatalog.Artifact, djspec *core.DynamicJobSpec) (storage.DataReference, error) {
	dataLocation, err := m.getFutureDataLocation(ctx, artifact)
	if err != nil {
		return "", errors.NewDataCatalogErrorf(codes.Internal, "Unable to generate data location %s, err %v", dataLocation.String(), err)
	}
	err = m.store.WriteProtobuf(ctx, dataLocation, storage.Options{}, djspec)
	if err != nil {
		return "", errors.NewDataCatalogErrorf(codes.Internal, "Unable to store future artifact data in location %s, err %v", dataLocation.String(), err)
	}

	return dataLocation, nil
}

// Retrieve the literal value of the ArtifactData from its specified location
func (m *artifactDataStore) GetData(ctx context.Context, dataModel models.ArtifactData) (*core.Literal, error) {
	var value core.Literal
	err := m.store.ReadProtobuf(ctx, storage.DataReference(dataModel.Location), &value)
	if err != nil {
		return nil, errors.NewDataCatalogErrorf(codes.Internal, "Unable to read artifact data from location %s, err %v", dataModel.Location, err)
	}

	return &value, nil
}

// Retrieve the future data from the future storage prefix
func (m *artifactDataStore) GetFutureData(ctx context.Context, dataModel models.ArtifactData) (*core.DynamicJobSpec, error) {
	var djSpec core.DynamicJobSpec
	err := m.store.ReadProtobuf(ctx, storage.DataReference(dataModel.Location), &djSpec)
	if err != nil {
		return nil, errors.NewDataCatalogErrorf(codes.Internal, "Unable to read future artifact data from location %s, err %v", dataModel.Location, err)
	}
	fmt.Printf("the dj spec: %+v\n", djSpec.GetNodes()[0])

	return &djSpec, nil
}

// DeleteData removes the stored artifact data from the underlying blob storage
func (m *artifactDataStore) DeleteData(ctx context.Context, dataModel models.ArtifactData) error {
	if err := m.store.Delete(ctx, storage.DataReference(dataModel.Location)); err != nil {
		return errors.NewDataCatalogErrorf(codes.Internal, "Unable to delete artifact data in location %s, err %v", dataModel.Location, err)
	}

	return nil
}

func NewArtifactDataStore(store *storage.DataStore, storagePrefix storage.DataReference) ArtifactDataStore {
	return &artifactDataStore{
		store:         store,
		storagePrefix: storagePrefix,
	}
}
