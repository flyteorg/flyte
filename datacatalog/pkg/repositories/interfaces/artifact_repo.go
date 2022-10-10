package interfaces

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

//go:generate mockery -name=ArtifactRepo -output=../mocks -case=underscore

type ArtifactRepo interface {
	Create(ctx context.Context, in models.Artifact) error
	Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error)
	List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error)
	Update(ctx context.Context, artifact models.Artifact) error
	Delete(ctx context.Context, artifact models.Artifact) error
}
