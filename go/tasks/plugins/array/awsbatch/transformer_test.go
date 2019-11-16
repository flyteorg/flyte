/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"
	"testing"

	flyteK8sConfig "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	mocks2 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/lyft/flyteplugins/go/tasks/plugins/array/awsbatch/config"

	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/stretchr/testify/assert"
)

func createSampleContainerTask() *core.Container {
	return &core.Container{
		Command: []string{"cmd"},
		Args:    []string{"{{$inputPrefix}}"},
		Image:   "img1",
		Config: []*core.KeyValuePair{
			{
				Key:   DynamicTaskQueueKey,
				Value: "child_queue",
			},
		},
	}
}

func ref(s string) *string {
	return &s
}

func TestArrayJobToBatchInput(t *testing.T) {
	expectedBatchInput := &batch.SubmitJobInput{
		ArrayProperties: &batch.ArrayProperties{
			Size: refInt(10),
		},
		JobDefinition: refStr(""),
		JobName:       refStr("Job_Name"),
		JobQueue:      refStr("child_queue"),
		ContainerOverrides: &batch.ContainerOverrides{
			Command: []*string{ref("cmd"), ref("/inputs/prefix")},
			Environment: []*batch.KeyValuePair{
				{Name: refStr("BATCH_JOB_ARRAY_INDEX_VAR_NAME"), Value: refStr("AWS_BATCH_JOB_ARRAY_INDEX")},
			},
			Memory: refInt(700),
			Vcpus:  refInt(2),
		},
	}

	input := &plugins.ArrayJob{
		Size:        10,
		Parallelism: 5,
	}

	id := &mocks.TaskExecutionID{}
	id.OnGetGeneratedName().Return("Job_Name")
	id.OnGetID().Return(core.TaskExecutionIdentifier{})

	to := &mocks.TaskOverrides{}
	to.OnGetConfig().Return(&v12.ConfigMap{
		Data: map[string]string{
			DynamicTaskQueueKey: "child_queue",
		},
	})

	tMetadata := &mocks.TaskExecutionMetadata{}
	tMetadata.OnGetAnnotations().Return(map[string]string{"aKey": "aVal"})
	tMetadata.OnGetNamespace().Return("ns")
	tMetadata.OnGetLabels().Return(map[string]string{"lKey": "lVal"})
	tMetadata.OnGetOwnerReference().Return(v1.OwnerReference{Name: "x"})
	tMetadata.OnGetTaskExecutionID().Return(id)
	tMetadata.OnGetOverrides().Return(to)

	ir := &mocks2.InputReader{}
	ir.OnGetInputPath().Return("inputs.pb")
	ir.OnGetInputPrefixPath().Return("/inputs/prefix")

	or := &mocks2.OutputWriter{}
	or.OnGetOutputPrefixPath().Return("/path/output")

	taskCtx := &mocks.TaskExecutionContext{}
	taskCtx.OnTaskExecutionMetadata().Return(tMetadata)
	taskCtx.OnInputReader().Return(ir)
	taskCtx.OnOutputWriter().Return(or)

	st, err := utils.MarshalObjToStruct(input)
	assert.NoError(t, err)

	taskTemplate := &core.TaskTemplate{
		Id:     &core.Identifier{Name: "Job_Name"},
		Custom: st,
		Target: &core.TaskTemplate_Container{
			Container: createSampleContainerTask(),
		},
	}

	tr := &mocks.TaskReader{}
	tr.OnReadMatch(mock.Anything).Return(taskTemplate, nil)

	taskCtx.OnTaskReader().Return(tr)

	ctx := context.Background()
	batchInput, err := FlyteTaskToBatchInput(ctx, taskCtx, "", &config.Config{})
	assert.NoError(t, err)

	batchInput = UpdateBatchInputForArray(ctx, batchInput, input.Size)
	assert.NotNil(t, batchInput)
	assert.Equal(t, *expectedBatchInput, *batchInput)
}

func Test_getEnvVarsForTask(t *testing.T) {
	ctx := context.Background()
	id := &mocks.TaskExecutionID{}
	id.OnGetGeneratedName().Return("Job_Name")
	id.OnGetID().Return(core.TaskExecutionIdentifier{})

	assert.NoError(t, flyteK8sConfig.SetK8sPluginConfig(&flyteK8sConfig.K8sPluginConfig{
		DefaultEnvVars: map[string]string{
			"MyKey": "BadVal",
		},
	}))

	envVars := getEnvVarsForTask(ctx, id, nil, map[string]string{
		"MyKey": "MyVal",
	})

	assert.Equal(t, []v12.EnvVar{
		{
			Name:  "MyKey",
			Value: "MyVal",
		},
	}, envVars)
}
