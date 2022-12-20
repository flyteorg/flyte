package interfaces

import (
	"context"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -name=Updater -case=underscore

type Updater interface {
	UpdateNamedEntity(ctx context.Context, name, project, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error
}
