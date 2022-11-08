package interfaces

import (
	"context"

	idl_datacatalog "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

//go:generate mockery -all -output=../mocks -case=underscore

type ArtifactManager interface {
	CreateArtifact(ctx context.Context, request *idl_datacatalog.CreateArtifactRequest) (*idl_datacatalog.CreateArtifactResponse, error)
	GetArtifact(ctx context.Context, request *idl_datacatalog.GetArtifactRequest) (*idl_datacatalog.GetArtifactResponse, error)
	ListArtifacts(ctx context.Context, request *idl_datacatalog.ListArtifactsRequest) (*idl_datacatalog.ListArtifactsResponse, error)
	UpdateArtifact(ctx context.Context, request *idl_datacatalog.UpdateArtifactRequest) (*idl_datacatalog.UpdateArtifactResponse, error)
}
