package server

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
)

type StorageInterface interface {
	CreateArtifact(context.Context, *models.Artifact) (models.Artifact, error)

	GetArtifact(ctx context.Context) (models.Artifact, error)
}
