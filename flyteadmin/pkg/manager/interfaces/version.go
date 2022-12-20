package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery -name VersionInterface -output=../mocks -case=underscore

// Interface for managing Flyte admin version
type VersionInterface interface {
	GetVersion(ctx context.Context, r *admin.GetVersionRequest) (*admin.GetVersionResponse, error)
}
