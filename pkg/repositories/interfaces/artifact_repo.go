package interfaces

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
)

type ArtifactRepo interface {
	Create(ctx context.Context, in models.Artifact) error
	Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error)
	List(ctx context.Context, datasetKey models.DatasetKey, in models.ListModelsInput) ([]models.Artifact, error)
}
