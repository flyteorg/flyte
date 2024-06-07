package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestRawOutputConfig(t *testing.T) {
	r := RawOutputDataConfig{&admin.RawOutputDataConfig{
		OutputLocationPrefix: "s3://bucket",
	}}
	assert.Equal(t, "s3://bucket", r.OutputLocationPrefix)
}

func TestWrappedRawOutputConfigDeepCopy(t *testing.T) {
	rawOutputDataConfig := RawOutputDataConfig{&admin.RawOutputDataConfig{
		OutputLocationPrefix: "s3://bucket",
	}}
	rawOutputDataConfigCopy := rawOutputDataConfig.DeepCopy()

	assert.True(t, rawOutputDataConfig.RawOutputDataConfig != rawOutputDataConfigCopy.RawOutputDataConfig)
	assert.True(t, proto.Equal(rawOutputDataConfig.RawOutputDataConfig, rawOutputDataConfigCopy.RawOutputDataConfig))
}

func TestExecutionConfigDeepCopy(t *testing.T) {
	// 1. Create an ExecutionConfig object (including the wrapper object - WorkflowExecutionIdentifier)
	interruptible := true
	executionConfig := &ExecutionConfig{
		TaskPluginImpls: map[string]TaskPluginOverride{},
		MaxParallelism:  32,
		RecoveryExecution: WorkflowExecutionIdentifier{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Org:     "organization",
				Name:    "name",
			},
		},
		TaskResources: TaskResources{
			Requests: TaskResourceSpec{
				CPU: resource.Quantity{
					Format: "1",
				},
				Memory: resource.Quantity{
					Format: "1Gi",
				},
			},
		},
		Interruptible:  &interruptible,
		OverwriteCache: false,
		EnvironmentVariables: map[string]string{
			"key": "value",
		},
	}

	// 2. Deep copy the wrapper object
	executionConfigCopy := executionConfig.DeepCopy()

	// 3. Assert that the pointers are different
	assert.True(t, executionConfig.RecoveryExecution.WorkflowExecutionIdentifier != executionConfigCopy.RecoveryExecution.WorkflowExecutionIdentifier)

	// 4. Assert that the values are the same
	assert.True(t, proto.Equal(executionConfig.RecoveryExecution.WorkflowExecutionIdentifier, executionConfigCopy.RecoveryExecution.WorkflowExecutionIdentifier))
	assert.Equal(t, executionConfig, executionConfigCopy)
}
