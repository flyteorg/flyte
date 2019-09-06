package catalog

import (
	"context"

	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NoOpDiscovery
type NoOpDiscovery struct{}

func (d *NoOpDiscovery) Get(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
	logger.Infof(ctx, "No-op Discovery Get invoked. Returning NotFound")
	return nil, status.Error(codes.NotFound, "No-op Discovery default behavior.")
}

func (d *NoOpDiscovery) Put(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
	logger.Infof(ctx, "No-op Discovery Put invoked. Doing nothing")
	return nil
}

func NewNoOpDiscovery() *NoOpDiscovery {
	return &NoOpDiscovery{}
}
