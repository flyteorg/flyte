package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

//go:generate mockery -name=ArtifactRepo -output=../mocks -case=underscore

type ArtifactRepo interface {
	Create(ctx context.Context, id *datacatalog.DatasetID, in models.Artifact) error
	Get(ctx context.Context, id *datacatalog.DatasetID, artifactID string) (models.Artifact, error)
	List(ctx context.Context, id *datacatalog.DatasetID, in models.ListModelsInput) ([]models.Artifact, error)
	Update(ctx context.Context, id *datacatalog.DatasetID, artifact models.Artifact) error
}
