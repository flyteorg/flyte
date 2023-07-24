/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package awsbatch

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/duration"

	"k8s.io/apimachinery/pkg/api/resource"

	flyteK8sConfig "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"

	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
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

func TestResourceRequirementsToBatchRequirements(t *testing.T) {
	memoryTests := []struct {
		Input    string
		Expected int64
	}{
		// resource Quantity gets the ceiling of values to the nearest scale
		{"200", 1},
		{"200M", 200},
		{"64G", 64000},
	}

	for i, testCase := range memoryTests {
		t.Run(fmt.Sprintf("Memory Test [%v] %v", i, testCase.Input), func(t *testing.T) {
			q, err := resource.ParseQuantity(testCase.Input)
			if assert.NoError(t, err) {
				assert.Equal(t, testCase.Expected, q.ScaledValue(resource.Mega),
					"Expected != Actual for test case [%v] with Input [%v]", i, testCase.Input)
			}
		})
	}

	cpuTests := []struct {
		Input    string
		Expected int64
	}{
		// resource Quantity gets the ceiling of values to the nearest scale
		{"200", 200},
		{"1M", 1000000},
		{"15000m", 15},
	}

	for i, testCase := range cpuTests {
		t.Run(fmt.Sprintf("CPU Test [%v] %v", i, testCase.Input), func(t *testing.T) {
			q, err := resource.ParseQuantity(testCase.Input)
			if assert.NoError(t, err) {
				assert.Equal(t, testCase.Expected, q.Value(),
					"Expected != Actual for test case [%v] with Input [%v]", i, testCase.Input)
			}
		})
	}
}

func TestToTimeout(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		timeout := toTimeout(nil, 0*time.Second)
		assert.Nil(t, timeout)
	})

	t.Run("TaskTemplate duration set", func(t *testing.T) {
		timeout := toTimeout(&duration.Duration{Seconds: 100}, 3*24*time.Hour)
		assert.NotNil(t, timeout.AttemptDurationSeconds)
		assert.Equal(t, int64(100), *timeout.AttemptDurationSeconds)
	})

	t.Run("Default timeout used", func(t *testing.T) {
		timeout := toTimeout(nil, 3*24*time.Hour)
		assert.NotNil(t, timeout.AttemptDurationSeconds)
		assert.Equal(t, int64((3 * 24 * time.Hour).Seconds()), *timeout.AttemptDurationSeconds)
	})

	t.Run("TaskTemplate duration set to 0", func(t *testing.T) {
		timeout := toTimeout(&duration.Duration{Seconds: 0}, 3*24*time.Hour)
		assert.NotNil(t, timeout.AttemptDurationSeconds)
		assert.Equal(t, int64((3 * 24 * time.Hour).Seconds()), *timeout.AttemptDurationSeconds)
	})
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
				{Name: ref(failOnError), Value: refStr("true")},
				{Name: refStr("BATCH_JOB_ARRAY_INDEX_VAR_NAME"), Value: refStr("AWS_BATCH_JOB_ARRAY_INDEX")},
			},
			Memory: refInt(1074),
			Vcpus:  refInt(1),
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

	to.OnGetResources().Return(&v12.ResourceRequirements{
		Limits:   v12.ResourceList{},
		Requests: v12.ResourceList{},
	})

	tMetadata := &mocks.TaskExecutionMetadata{}
	tMetadata.OnGetAnnotations().Return(map[string]string{"aKey": "aVal"})
	tMetadata.OnGetNamespace().Return("ns")
	tMetadata.OnGetLabels().Return(map[string]string{"lKey": "lVal"})
	tMetadata.OnGetOwnerReference().Return(v1.OwnerReference{Name: "x"})
	tMetadata.OnGetTaskExecutionID().Return(id)
	tMetadata.OnGetOverrides().Return(to)
	tMetadata.OnGetPlatformResources().Return(&v12.ResourceRequirements{})

	ir := &mocks2.InputReader{}
	ir.OnGetInputPath().Return("inputs.pb")
	ir.OnGetInputPrefixPath().Return("/inputs/prefix")
	ir.OnGetMatch(mock.Anything).Return(nil, nil)

	or := &mocks2.OutputWriter{}
	or.OnGetOutputPrefixPath().Return("/path/output")
	or.OnGetRawOutputPrefix().Return("s3://")
	or.OnGetCheckpointPrefix().Return("/checkpoint")
	or.OnGetPreviousCheckpointsPrefix().Return("/prev")

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
		Type: arrayTaskType,
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

	taskTemplate.Type = array.AwsBatchTaskType
	tr.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
	taskCtx.OnTaskReader().Return(tr)

	ctx = context.Background()
	_, err = FlyteTaskToBatchInput(ctx, taskCtx, "", &config.Config{})
	assert.NoError(t, err)
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

	expected := map[string]string{
		"FLYTE_FAIL_ON_ERROR": "true",
		"MyKey":               "MyVal",
	}

	assert.Len(t, envVars, len(expected))
	for _, envVar := range envVars {
		assert.Contains(t, expected, envVar.Name)
		assert.Equal(t, expected[envVar.Name], envVar.Value)
	}
}
