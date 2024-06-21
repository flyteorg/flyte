package controller

import (
	"context"

	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/unionai/flyte/fasttask/plugin"

	pluginscore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// TODO currently we only have one ExecutionEnvironment builder (ie. fast-task), but in the future
// we plan to have more. this code should be minorly refactored to reflect that support.
type ExecutionEnvClient struct {
	envBuilders map[string]*plugin.InMemoryEnvBuilder
}

// Create initializes an execution environment with the given ID and spec.
func (e *ExecutionEnvClient) Create(ctx context.Context, executionEnvID string, executionEnvSpec *_struct.Struct) (*_struct.Struct, error) {
	fastTaskEnvBuilder := e.envBuilders["fast-task"]
	return fastTaskEnvBuilder.Create(ctx, executionEnvID, executionEnvSpec)
}

// Get returns the execution environment with the given ID.
func (e *ExecutionEnvClient) Get(ctx context.Context, executionEnvID string) *_struct.Struct {
	fastTaskEnvBuilder := e.envBuilders["fast-task"]
	return fastTaskEnvBuilder.Get(ctx, executionEnvID)
}

// Status returns the status of the execution environment with the given ID.
func (e *ExecutionEnvClient) Status(ctx context.Context, executionEnvID string) (interface{}, error) {
	fastTaskEnvBuilder := e.envBuilders["fast-task"]
	return fastTaskEnvBuilder.Status(ctx, executionEnvID)
}

// NewExecutionEnvClient creates a new ExecutionEnvClient.
func NewExecutionEnvClient(ctx context.Context, kubeClient executors.Client, scope promutils.Scope) (pluginscore.ExecutionEnvClient, error) {
	envBuilderScope := scope.NewSubScope("env_builder")

	fastTaskEnvBuilder := plugin.NewEnvironmentBuilder(kubeClient, envBuilderScope.NewSubScope("fast_task"))
	if err := fastTaskEnvBuilder.Start(ctx); err != nil {
		return nil, err
	}

	return &ExecutionEnvClient{
		envBuilders: map[string]*plugin.InMemoryEnvBuilder{
			"fast-task": fastTaskEnvBuilder,
		},
	}, nil
}
