package executionenv

import (
	"context"
	"fmt"

	_struct "github.com/golang/protobuf/ptypes/struct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/unionai/fasttask/plugin"
)

type ExecutionEnvironmentClient struct {
	environments map[string]*_struct.Struct
	fastTaskEnvBuilder *plugin.EnvironmentBuilder // TODO @hamersaw - this is a hack
}

func (e *ExecutionEnvironmentClient) GetEnvironment(executionID *core.WorkflowExecutionIdentifier, executionEnvID string) *_struct.Struct {
	id := buildID(executionID, executionEnvID)
	fmt.Printf("HAMERSAW - GetEnvironment '%s'\n", id)

	env, _ := e.environments[id]
	return env
}

func (e *ExecutionEnvironmentClient) CreateEnvironment(ctx context.Context, ownerReference metav1.OwnerReference, executionID *core.WorkflowExecutionIdentifier,
	executionEnvID string, executionEnvSpec *_struct.Struct) error {

	id := buildID(executionID, executionEnvID)
	fmt.Printf("HAMERSAW - CreateEnvironment '%s'\n", id)

	if _, exists := e.environments[id]; exists {
		// TODO log creating an existing environment
		return nil
	}

	env, err := e.fastTaskEnvBuilder.CreateEnvironment(ctx, ownerReference, executionID, executionEnvID, executionEnvSpec)
	if err != nil {
		return err
	}

	e.environments[id] = env
	return nil
}

func (e *ExecutionEnvironmentClient) DeleteEnvironment(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	executionEnvID string, executionEnvSpec *_struct.Struct) error {

	id := buildID(executionID, executionEnvID)
	fmt.Printf("HAMERSAW - DeleteEnvironment '%s'\n", id)

	if err := e.fastTaskEnvBuilder.DeleteEnvironment(ctx, executionID, executionEnvID, executionEnvSpec); err != nil {
		return err
	}

	delete(e.environments, id)
	return nil
}

func (e *ExecutionEnvironmentClient) ConfirmEnvironment(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	executionEnvID string, executionEnvSpec *_struct.Struct) error {

	fmt.Printf("HAMERSAW - TODO ConfirmEnvironment '%s'\n", executionEnvID)
	return nil
}

func NewExecutionEnvironmentClient(_ context.Context, kubeClient executors.Client, scope promutils.Scope) *ExecutionEnvironmentClient {
	return &ExecutionEnvironmentClient{
		environments: make(map[string]*_struct.Struct),
		fastTaskEnvBuilder: plugin.NewEnvironmentBuilder(kubeClient),
	}
}

func buildID(executionID *core.WorkflowExecutionIdentifier, executionEnvID string) string {
	return fmt.Sprintf("%s:%s:%s:%s", executionID.Project, executionID.Domain, executionID.Name, executionEnvID)
}
