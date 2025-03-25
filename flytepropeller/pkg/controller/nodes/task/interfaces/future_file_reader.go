package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

//go:generate mockery-v2 --name FutureFileReaderInterface --output ../mocks --case=snake --with-expecter

type FutureFileReaderInterface interface {
	GetLoc() storage.DataReference
	Exists(ctx context.Context) (bool, error)
	Read(ctx context.Context) (*core.DynamicJobSpec, error)
	CacheExists(ctx context.Context) (bool, error)
	Cache(ctx context.Context, wf *v1alpha1.FlyteWorkflow, workflowClosure *core.CompiledWorkflowClosure) error
	RetrieveCache(ctx context.Context) (task.CacheContents, error)
}
