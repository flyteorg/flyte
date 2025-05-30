package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
)

//go:generate mockery --name=ArtifactRepo --output=../mocks --case=underscore --with-expecter

type ArtifactRepo interface {
	Create(ctx context.Context, in models.Artifact) error
	GetAndFilterExpired(ctx context.Context, in models.ArtifactKey) (models.Artifact, error)
	ListAndFilterExpired(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error)
	Update(ctx context.Context, artifact models.Artifact) error
}
