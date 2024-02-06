package executionenv

import (
	"context"
	"fmt"

	_struct "github.com/golang/protobuf/ptypes/struct"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type ExecutionEnvironmentClient struct {
}

func (e *ExecutionEnvironmentClient) GetEnvironment(nCtx interfaces.NodeExecutionContext, executionEnvID string) *_struct.Struct {
	fmt.Printf("HAMERSAW - GetEnvironment for node '%s'\n", nCtx.NodeID())
	return nil
}

func (e *ExecutionEnvironmentClient) CreateEnvironment(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	executionEnvID string, executionEnvSpec *_struct.Struct) error {

	fmt.Printf("HAMERSAW - CreateEnvironment '%s'\n", executionEnvID)
	/*switch executionEnvSpec.Type {
	case core.EnvironmentType_FAST_TASK:
		return fmt.Errorf("not implemented")
	default:
		return fmt.Errorf("not implemented")
	}*/

	return nil
}

func (e *ExecutionEnvironmentClient) DeleteEnvironment(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	executionEnvID string, executionEnvSpec *_struct.Struct) error {

	fmt.Printf("HAMERSAW - DeleteEnvironment '%s'\n", executionEnvID)
	/*switch executionEnvSpec.Type {
	case core.EnvironmentType_FAST_TASK:
		return fmt.Errorf("not implemented")
	default:
		return fmt.Errorf("not implemented")
	}*/

	return nil
}

func (e *ExecutionEnvironmentClient) ConfirmEnvironment(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	executionEnvID string, executionEnvSpec *_struct.Struct) error {

	fmt.Printf("HAMERSAW - ConfirmEnvironment '%s'\n", executionEnvID)
	/*switch executionEnvSpec.Type {
	case core.EnvironmentType_FAST_TASK:
		return fmt.Errorf("not implemented")
	default:
		return fmt.Errorf("not implemented")
	}*/

	return nil
}

func NewExecutionEnvironmentClient(_ context.Context, kubeClient executors.Client, scope promutils.Scope) *ExecutionEnvironmentClient {
	return &ExecutionEnvironmentClient{}
}
