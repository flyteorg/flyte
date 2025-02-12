package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery-v2 --name=VersionInterface --output=../mocks --case=underscore --with-expecter

// Interface for managing Flyte admin version
type VersionInterface interface {
	GetVersion(ctx context.Context, r *admin.GetVersionRequest) (*admin.GetVersionResponse, error)
}
