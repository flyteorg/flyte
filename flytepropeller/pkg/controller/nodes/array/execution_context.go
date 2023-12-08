package array

import (
	"strconv"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
)

const (
	FlyteK8sArrayIndexVarName string = "FLYTE_K8S_ARRAY_INDEX"
	JobIndexVarName           string = "BATCH_JOB_ARRAY_INDEX_VAR_NAME"
)

type arrayExecutionContext struct {
	executors.ExecutionContext
	executionConfig v1alpha1.ExecutionConfig
}

func (a *arrayExecutionContext) GetExecutionConfig() v1alpha1.ExecutionConfig {
	return a.executionConfig
}

func newArrayExecutionContext(executionContext executors.ExecutionContext, subNodeIndex int) *arrayExecutionContext {
	executionConfig := executionContext.GetExecutionConfig()
	if executionConfig.EnvironmentVariables == nil {
		executionConfig.EnvironmentVariables = make(map[string]string)
	}
	executionConfig.EnvironmentVariables[JobIndexVarName] = FlyteK8sArrayIndexVarName
	executionConfig.EnvironmentVariables[FlyteK8sArrayIndexVarName] = strconv.Itoa(subNodeIndex)
	executionConfig.MaxParallelism = 0 // hardcoded to 0 because parallelism is handled by the array node

	return &arrayExecutionContext{
		ExecutionContext: executionContext,
		executionConfig:  executionConfig,
	}
}
