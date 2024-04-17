package core

import (
	"context"
	coreIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate mockery -all -output=./mocks -case=underscore

type ConnectionManager interface {
	Get(ctx context.Context, key string) (*coreIdl.Connection, error)
}
