package interfaces

import (
	"context"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type Updater interface {
	UpdateNamedEntity(ctx context.Context, name, project, domain string, rsType core.ResourceType, cmdCtx cmdCore.CommandContext) error
}
