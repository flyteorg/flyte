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

	GetArtifact(context.Context, core.ArtifactQuery) (models.Artifact, error)
}

type BlobStoreInterface interface {
	OffloadArtifactCard(ctx context.Context, name string, version string, userMetadata *any.Any) (stdLibStorage.DataReference, error)

	RetrieveArtifactCard(context.Context, stdLibStorage.DataReference) (*any.Any, error)
}

// This interface is under server/ because it's going to take a pointer to
// the main service itself to be able to directly invoke the endpoint.
// If we're okay with a local grpc call, we can move it a level higher.
