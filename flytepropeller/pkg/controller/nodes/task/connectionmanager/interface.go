package connectionmanager

import (
	"context"
	flyteidl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// ConnectionManager is an interface that allows retrieving a connection for a task.
type ConnectionManager interface {
	Get(ctx context.Context, key string) (flyteidl.Connection, error)
}
