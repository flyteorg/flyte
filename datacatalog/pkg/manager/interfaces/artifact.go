package interfaces

import (
	"context"

	idl_datacatalog "github.com/lyft/datacatalog/protos/gen"
)

type ArtifactManager interface {
	CreateArtifact(ctx context.Context, request idl_datacatalog.CreateArtifactRequest) (*idl_datacatalog.CreateArtifactResponse, error)
	GetArtifact(ctx context.Context, request idl_datacatalog.GetArtifactRequest) (*idl_datacatalog.GetArtifactResponse, error)
}
