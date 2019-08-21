package impl

import (
	"context"

	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"
	datacatalog "github.com/lyft/datacatalog/protos/gen"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/transformers"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

type artifactManager struct {
	repo          repositories.RepositoryInterface
	artifactStore ArtifactDataStore
}

// Create an Artifact along with the associated ArtifactData. The ArtifactData will be stored in an offloaded location.
func (m *artifactManager) CreateArtifact(ctx context.Context, request datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error) {
	artifact := request.Artifact
	err := validators.ValidateArtifact(artifact)
	if err != nil {
		return nil, err
	}

	datasetKey := transformers.FromDatasetID(*artifact.Dataset)

	// The dataset must exist for the artifact, let's verify that first
	_, err = m.repo.DatasetRepo().Get(ctx, datasetKey)
	if err != nil {
		return nil, err
	}

	// create Artifact Data offloaded storage files
	artifactDataModels := make([]models.ArtifactData, len(request.Artifact.Data))
	for i, artifactData := range request.Artifact.Data {
		dataLocation, err := m.artifactStore.PutData(ctx, *artifact, *artifactData)
		if err != nil {
			return nil, err
		}

		artifactDataModels[i].Name = artifactData.Name
		artifactDataModels[i].Location = dataLocation.String()
	}

	artifactModel, err := transformers.CreateArtifactModel(request, artifactDataModels)
	if err != nil {
		return nil, err
	}

	err = m.repo.ArtifactRepo().Create(ctx, artifactModel)
	if err != nil {
		return nil, err
	}
	return &datacatalog.CreateArtifactResponse{}, nil
}

// Get the Artifact and its associated ArtifactData. The request can query by ArtifactID or TagName.
func (m *artifactManager) GetArtifact(ctx context.Context, request datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error) {
	datasetID := request.Dataset
	err := validators.ValidateGetArtifactRequest(request)
	if err != nil {
		return nil, err
	}

	var artifactModel models.Artifact
	switch request.QueryHandle.(type) {
	case *datacatalog.GetArtifactRequest_ArtifactId:
		artifactKey := transformers.ToArtifactKey(*datasetID, request.GetArtifactId())
		artifactModel, err = m.repo.ArtifactRepo().Get(ctx, artifactKey)

		if err != nil {
			return nil, err
		}
	case *datacatalog.GetArtifactRequest_TagName:
		tagKey := transformers.ToTagKey(*datasetID, request.GetTagName())
		tag, err := m.repo.TagRepo().Get(ctx, tagKey)

		if err != nil {
			return nil, err
		}

		artifactModel = tag.Artifact
	}

	if len(artifactModel.ArtifactData) == 0 {
		return nil, errors.NewDataCatalogErrorf(codes.Internal, "artifact [%+v] does not have artifact data associated", request)
	}

	artifact, err := transformers.FromArtifactModel(artifactModel)
	if err != nil {
		return nil, err
	}

	artifactDataList := make([]*datacatalog.ArtifactData, len(artifactModel.ArtifactData))
	for i, artifactData := range artifactModel.ArtifactData {
		value, err := m.artifactStore.GetData(ctx, artifactData)
		if err != nil {
			return nil, err
		}

		artifactDataList[i] = &datacatalog.ArtifactData{
			Name:  artifactData.Name,
			Value: value,
		}
	}
	artifact.Data = artifactDataList

	return &datacatalog.GetArtifactResponse{
		Artifact: &artifact,
	}, nil
}

func NewArtifactManager(repo repositories.RepositoryInterface, store *storage.DataStore, storagePrefix storage.DataReference, artifactScope promutils.Scope) interfaces.ArtifactManager {
	return &artifactManager{
		repo:          repo,
		artifactStore: NewArtifactDataStore(store, storagePrefix),
	}
}
