package catalog

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
)

//go:generate mockery -all -case=underscore

// Metadata to be associated with the catalog object
type Metadata struct {
	WorkflowExecutionIdentifier *core.WorkflowExecutionIdentifier
	NodeExecutionIdentifier     *core.NodeExecutionIdentifier
	TaskExecutionIdentifier     *core.TaskExecutionIdentifier
}

// An identifier for a catalog object.
type Key struct {
	Identifier     core.Identifier
	CacheVersion   string
	TypedInterface core.TypedInterface
	InputReader    io.InputReader
}

func (k Key) String() string {
	return fmt.Sprintf("%v:%v", k.Identifier, k.CacheVersion)
}

// Default Catalog client that allows memoization and indexing of intermediate data in Flyte
type Client interface {
	Get(ctx context.Context, key Key) (io.OutputReader, error)
	Put(ctx context.Context, key Key, reader io.OutputReader, metadata Metadata) error
}

func IsNotFound(err error) bool {
	taskStatus, ok := status.FromError(err)
	return ok && taskStatus.Code() == codes.NotFound
}
