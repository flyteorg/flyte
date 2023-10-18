package server

import (
	"context"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	stdLibStorage "github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/golang/protobuf/ptypes/any"
)

type StorageInterface interface {
	CreateArtifact(context.Context, models.Artifact) (models.Artifact, error)

	GetArtifact(context.Context, core.ArtifactQuery, bool) (models.Artifact, error)
}

type BlobStoreInterface interface {
	OffloadArtifactCard(ctx context.Context, name string, version string, userMetadata *any.Any) (stdLibStorage.DataReference, error)

	RetrieveArtifactCard(context.Context, stdLibStorage.DataReference) (*any.Any, error)
}
