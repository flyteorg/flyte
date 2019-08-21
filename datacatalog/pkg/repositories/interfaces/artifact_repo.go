package interfaces

import (
	"context"

	"github.com/lyft/datacatalog/pkg/repositories/models"
)

type ArtifactRepo interface {
	Create(ctx context.Context, in models.Artifact) error
	Get(ctx context.Context, in models.ArtifactKey) (models.Artifact, error)
}
