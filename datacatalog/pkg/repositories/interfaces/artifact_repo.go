package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
)

//go:generate mockery-v2 --name=ArtifactRepo --output=../mocks --case=underscore --with-expecter

type ArtifactRepo interface {
	Create(ctx context.Context, in models.Artifact) error
	Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error)
	List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error)
	Update(ctx context.Context, artifact models.Artifact) error
}
