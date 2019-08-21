package impl

import (
	"context"

	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/transformers"
	datacatalog "github.com/lyft/datacatalog/protos/gen"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
)

type tagManager struct {
	repo  repositories.RepositoryInterface
	store *storage.DataStore
}

func (m *tagManager) AddTag(ctx context.Context, request datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error) {

	if err := validators.ValidateTag(request.Tag); err != nil {
		return nil, err
	}

	// verify the artifact exists before adding a tag to it
	datasetID := *request.Tag.Dataset
	artifactKey := transformers.ToArtifactKey(datasetID, request.Tag.ArtifactId)
	_, err := m.repo.ArtifactRepo().Get(ctx, artifactKey)
	if err != nil {
		return nil, err
	}

	tagKey := transformers.ToTagKey(datasetID, request.Tag.Name)
	err = m.repo.TagRepo().Create(ctx, models.Tag{
		TagKey:     tagKey,
		ArtifactID: request.Tag.ArtifactId,
	})
	if err != nil {
		return nil, err
	}

	return &datacatalog.AddTagResponse{}, nil
}

func NewTagManager(repo repositories.RepositoryInterface, store *storage.DataStore, tagScope promutils.Scope) interfaces.TagManager {
	return &tagManager{
		repo:  repo,
		store: store,
	}
}
