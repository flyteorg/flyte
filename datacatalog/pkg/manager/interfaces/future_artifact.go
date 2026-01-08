package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

type FutureArtifactManager interface {
	CreateFutureArtifact(ctx context.Context, request *datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error)
	GetFutureArtifact(ctx context.Context, request *datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error)
	UpdateFutureArtifact(ctx context.Context, request *datacatalog.UpdateArtifactRequest) (*datacatalog.UpdateArtifactResponse, error)
}
