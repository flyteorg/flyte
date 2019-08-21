package impl

import (
	"context"

	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"
	"github.com/lyft/datacatalog/pkg/repositories/transformers"
	datacatalog "github.com/lyft/datacatalog/protos/gen"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
)

type datasetManager struct {
	repo  repositories.RepositoryInterface
	store *storage.DataStore
}

// Create a Dataset with optional metadata. If one already exists a grpc AlreadyExists err will be returned
func (dm *datasetManager) CreateDataset(ctx context.Context, request datacatalog.CreateDatasetRequest) (*datacatalog.CreateDatasetResponse, error) {
	err := validators.ValidateDatasetID(request.Dataset.Id)
	if err != nil {
		return nil, err
	}

	datasetModel, err := transformers.CreateDatasetModel(request.Dataset)
	if err != nil {
		return nil, err
	}

	err = dm.repo.DatasetRepo().Create(ctx, *datasetModel)
	if err != nil {
		return nil, err
	}

	return &datacatalog.CreateDatasetResponse{}, nil
}

// Get a Dataset with the given DatasetID if it exists. If none exist a grpc NotFound err will be returned
func (dm *datasetManager) GetDataset(ctx context.Context, request datacatalog.GetDatasetRequest) (*datacatalog.GetDatasetResponse, error) {
	err := validators.ValidateDatasetID(request.Dataset)
	if err != nil {
		return nil, err
	}

	datasetKey := transformers.FromDatasetID(*request.Dataset)
	datasetModel, err := dm.repo.DatasetRepo().Get(ctx, datasetKey)

	if err != nil {
		return nil, err
	}

	datasetResponse, err := transformers.FromDatasetModel(datasetModel)
	if err != nil {
		return nil, err
	}

	return &datacatalog.GetDatasetResponse{
		Dataset: datasetResponse,
	}, nil
}

func NewDatasetManager(repo repositories.RepositoryInterface, store *storage.DataStore, datasetScope promutils.Scope) interfaces.DatasetManager {
	return &datasetManager{
		repo:  repo,
		store: store,
	}
}
