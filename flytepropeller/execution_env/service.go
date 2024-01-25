package executionenv

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type ExecutionEnvironmentService struct {
}

func (e *ExecutionEnvironmentService) GetEnvironment(nCtx interfaces.NodeExecutionContext, executionEnvID string) *core.ExecutionEnvironment {
	fmt.Printf("HAMERSAW - GetEnvironment for node '%s'\n", nCtx.NodeID())
	return nil
}

func (e *ExecutionEnvironmentService) CreateEnvironment(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	executionEnvID string, executionEnvSpec *core.ExecutionEnvironmentSpec) error {

	switch executionEnvSpec.Type {
	case core.EnvironmentType_FAST_TASK:
		return fmt.Errorf("not implemented")
	default:
		return fmt.Errorf("not implemented")
	}

	return nil
}

func (e *ExecutionEnvironmentService) DeleteEnvironment(ctx context.Context, nCtx interfaces.NodeExecutionContext, executionEnvID string) error {

	return fmt.Errorf("not implemented")
}

func (e *ExecutionEnvironmentService) ConfirmEnvironment(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	executionEnvID string, executionEnvSpec *core.ExecutionEnvironmentSpec) error {

	switch executionEnvSpec.Type {
	case core.EnvironmentType_FAST_TASK:
		return fmt.Errorf("not implemented")
	default:
		return fmt.Errorf("not implemented")
	}

	return nil
}

func NewExecutionEnvironmentService(_ context.Context, kubeClient executors.Client, scope promutils.Scope) *ExecutionEnvironmentService {
	return &ExecutionEnvironmentService{}
}
