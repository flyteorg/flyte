package array

import (
	"strconv"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
)

const (
	FlyteK8sArrayIndexVarName string = "FLYTE_K8S_ARRAY_INDEX"
	JobIndexVarName           string = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
)

type arrayExecutionContext struct {
	executors.ExecutionContext
	executionConfig    v1alpha1.ExecutionConfig
	currentParallelism *uint32
}

func (a *arrayExecutionContext) GetExecutionConfig() v1alpha1.ExecutionConfig {
	return a.executionConfig
}

func (a *arrayExecutionContext) CurrentParallelism() uint32 {
	return *a.currentParallelism
}

func (a *arrayExecutionContext) IncrementParallelism() uint32 {
	*a.currentParallelism = *a.currentParallelism + 1
	return *a.currentParallelism
}

func newArrayExecutionContext(executionContext executors.ExecutionContext, subNodeIndex int, currentParallelism *uint32, maxParallelism uint32) *arrayExecutionContext {
	executionConfig := executionContext.GetExecutionConfig()
	if executionConfig.EnvironmentVariables == nil {
		executionConfig.EnvironmentVariables = make(map[string]string)
	}
	executionConfig.EnvironmentVariables[JobIndexVarName] = FlyteK8sArrayIndexVarName
	executionConfig.EnvironmentVariables[FlyteK8sArrayIndexVarName] = strconv.Itoa(subNodeIndex)
	executionConfig.MaxParallelism = maxParallelism

	return &arrayExecutionContext{
		ExecutionContext:   executionContext,
		executionConfig:    executionConfig,
		currentParallelism: currentParallelism,
	}
}
