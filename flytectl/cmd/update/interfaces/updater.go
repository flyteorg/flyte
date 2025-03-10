package interfaces

import (
	"context"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery-v2 --name=Updater --case=underscore --with-expecter

type Updater interface {
	UpdateNamedEntity(ctx context.Context, name, project, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error
}
