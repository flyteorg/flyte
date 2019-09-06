package catalog

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"
)

type MockCatalogClient struct {
	GetFunc func(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error)
	PutFunc func(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error
}

func (m *MockCatalogClient) Get(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
	return m.GetFunc(ctx, task, inputPath)
}

func (m *MockCatalogClient) Put(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
	return m.PutFunc(ctx, task, execID, inputPath, outputPath)
}
