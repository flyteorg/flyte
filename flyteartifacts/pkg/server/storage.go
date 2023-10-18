package server

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type StorageInterface interface {
	CreateArtifact(context.Context, models.Artifact) (models.Artifact, error)

	GetArtifact(context.Context, core.ArtifactQuery, bool) (models.Artifact, error)
}
