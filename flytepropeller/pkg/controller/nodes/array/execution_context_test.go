package array

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
)

// mockImmutableExecutionContext implements just GetExecutionConfig for testing
type mockImmutableExecutionContext struct {
	executors.ImmutableExecutionContext
	config v1alpha1.ExecutionConfig
}

func (m *mockImmutableExecutionContext) GetExecutionConfig() v1alpha1.ExecutionConfig {
	return m.config
}

func TestNewArrayExecutionContext(t *testing.T) {
	baseConfig := v1alpha1.ExecutionConfig{
		MaxParallelism:       10,
		EnvironmentVariables: map[string]string{"EXISTING": "value"},
	}
	immExecContext := &mockImmutableExecutionContext{config: baseConfig}

	t.Run("MaxParallelismDisabled", func(t *testing.T) {
		parentExecContext := executors.NewExecutionContext(immExecContext, nil, nil, nil, executors.InitializeControlFlow())
		arrayExecContext := newArrayExecutionContext(parentExecContext, 0)

		assert.Equal(t, uint32(0), arrayExecContext.GetExecutionConfig().MaxParallelism)
	})

	t.Run("EnvironmentVariables", func(t *testing.T) {
		parentExecContext := executors.NewExecutionContext(immExecContext, nil, nil, nil, executors.InitializeControlFlow())
		arrayExecContext := newArrayExecutionContext(parentExecContext, 3)

		config := arrayExecContext.GetExecutionConfig()
		assert.Equal(t, "value", config.EnvironmentVariables["EXISTING"])
		assert.Equal(t, FlyteK8sArrayIndexVarName, config.EnvironmentVariables[JobIndexVarName])
		assert.Equal(t, "3", config.EnvironmentVariables[FlyteK8sArrayIndexVarName])

		// verify parent env vars not mutated
		assert.NotContains(t, baseConfig.EnvironmentVariables, JobIndexVarName)
	})

	t.Run("SubNodeParallelismIsolatedFromParent", func(t *testing.T) {
		parentControlFlow := executors.InitializeControlFlow()
		parentExecContext := executors.NewExecutionContext(immExecContext, nil, nil, nil, parentControlFlow)

		// simulate what buildArrayNodeContext does: create a fresh control flow for the subnode
		childControlFlow := executors.InitializeControlFlow()
		childBaseContext := executors.NewExecutionContext(parentExecContext, nil, nil, nil, childControlFlow)
		arrayExecContext := newArrayExecutionContext(childBaseContext, 0)

		// incrementing parallelism on the array execution context should not affect the parent
		arrayExecContext.IncrementParallelism()
		arrayExecContext.IncrementParallelism()
		arrayExecContext.IncrementParallelism()

		assert.Equal(t, uint32(3), arrayExecContext.CurrentParallelism())
		assert.Equal(t, uint32(0), parentControlFlow.CurrentParallelism())
	})
}