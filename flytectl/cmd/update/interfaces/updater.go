package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
)

//go:generate mockery -name=Updater -case=underscore

type Updater interface {
	UpdateNamedEntity(ctx context.Context, name, project, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error
}
