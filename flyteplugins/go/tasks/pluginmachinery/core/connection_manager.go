package core

import (
	"context"

	flyteidl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type ConnectionManager interface {
	Get(ctx context.Context, key string) (flyteidl.Connection, error)
}
