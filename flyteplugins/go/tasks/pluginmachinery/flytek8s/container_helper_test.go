package flytek8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var zeroQuantity = resource.MustParse("0")

func TestAssignResource(t *testing.T) {
	t.Run("Leave valid requests and limits unchanged", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("1"), resource.MustParse("2"),
			resource.MustParse("10"), resource.MustParse("20"))
		assert.True(t, res.Request.Equal(resource.MustParse("1")))
		assert.True(t, res.Limit.Equal(resource.MustParse("2")))
	})
	t.Run("Assign unset Request from Limit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			zeroQuantity, resource.MustParse("2"),
			resource.MustParse("10"), resource.MustParse("20"))
		assert.True(t, res.Request.Equal(resource.MustParse("2")))
		assert.True(t, res.Limit.Equal(resource.MustParse("2")))
	})
	t.Run("Assign unset Limit from Request", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("2"), zeroQuantity,
			resource.MustParse("10"), resource.MustParse("20"))
		assert.Equal(t, resource.MustParse("2"), res.Request)
		assert.Equal(t, resource.MustParse("2"), res.Limit)
	})
	t.Run("Assign from platform defaults", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			zeroQuantity, zeroQuantity,
			resource.MustParse("10"), resource.MustParse("20"))
		assert.Equal(t, resource.MustParse("10"), res.Request)
		assert.Equal(t, resource.MustParse("10"), res.Limit)
	})
	t.Run("Adjust Limit when Request > Limit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("10"), resource.MustParse("2"),
			resource.MustParse("10"), resource.MustParse("20"))
		assert.Equal(t, resource.MustParse("2"), res.Request)
		assert.Equal(t, resource.MustParse("2"), res.Limit)
	})
	t.Run("Adjust Limit > platformLimit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("1"), resource.MustParse("40"),
			resource.MustParse("10"), resource.MustParse("20"))
		assert.True(t, res.Request.Equal(resource.MustParse("1")))
		assert.True(t, res.Limit.Equal(resource.MustParse("20")))
	})
	t.Run("Adjust Request, Limit > platformLimit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("40"), resource.MustParse("50"),
			resource.MustParse("10"), resource.MustParse("20"))
		assert.True(t, res.Request.Equal(resource.MustParse("20")))
		assert.True(t, res.Limit.Equal(resource.MustParse("20")))
	})
}

func TestValidateResource(t *testing.T) {
	platformLimit := resource.MustParse("5")
	t.Run("adjust when Request > Limit", func(t *testing.T) {
		res := ensureResourceRange(resource.MustParse("4"), resource.MustParse("3"), platformLimit)
		assert.True(t, res.Request.Equal(resource.MustParse("3")))
		assert.True(t, res.Limit.Equal(resource.MustParse("3")))
	})
	t.Run("adjust when Request > platformLimit", func(t *testing.T) {
		res := ensureResourceRange(resource.MustParse("6"), platformLimit, platformLimit)
		assert.True(t, res.Request.Equal(platformLimit))
		assert.True(t, res.Limit.Equal(platformLimit))
	})
	t.Run("adjust when Limit > platformLimit", func(t *testing.T) {
		res := ensureResourceRange(resource.MustParse("4"), resource.MustParse("6"), platformLimit)
		assert.True(t, res.Request.Equal(resource.MustParse("4")))
		assert.True(t, res.Limit.Equal(platformLimit))
	})
	t.Run("nothing to do", func(t *testing.T) {
		res := ensureResourceRange(resource.MustParse("1"), resource.MustParse("2"), platformLimit)
		assert.True(t, res.Request.Equal(resource.MustParse("1")))
		assert.True(t, res.Limit.Equal(resource.MustParse("2")))
	})
}

func TestApplyResourceOverrides_OverrideCpu(t *testing.T) {
	platformRequirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("3"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU: resource.MustParse("10"),
		},
	}
	cpuRequest := resource.MustParse("1")
	overrides := ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: cpuRequest,
		},
	}, platformRequirements, assignIfUnset)
	assert.EqualValues(t, cpuRequest, overrides.Requests[v1.ResourceCPU])
	assert.EqualValues(t, cpuRequest, overrides.Limits[v1.ResourceCPU])

	cpuLimit := resource.MustParse("2")
	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: cpuRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU: cpuLimit,
		},
	}, platformRequirements, assignIfUnset)
	assert.EqualValues(t, cpuRequest, overrides.Requests[v1.ResourceCPU])
	assert.EqualValues(t, cpuLimit, overrides.Limits[v1.ResourceCPU])

	// Request equals Limit if not set
	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU: cpuLimit,
		},
	}, platformRequirements, assignIfUnset)
	assert.EqualValues(t, cpuLimit, overrides.Requests[v1.ResourceCPU])
	assert.EqualValues(t, cpuLimit, overrides.Limits[v1.ResourceCPU])
}

func TestApplyResourceOverrides_OverrideMemory(t *testing.T) {
	memoryRequest := resource.MustParse("1")
	platformRequirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("3"),
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: resource.MustParse("10"),
		},
	}
	overrides := ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: memoryRequest,
		},
	}, platformRequirements, assignIfUnset)
	assert.EqualValues(t, memoryRequest, overrides.Requests[v1.ResourceMemory])
	assert.EqualValues(t, memoryRequest, overrides.Limits[v1.ResourceMemory])

	memoryLimit := resource.MustParse("2")
	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: memoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: memoryLimit,
		},
	}, platformRequirements, assignIfUnset)
	assert.EqualValues(t, memoryRequest, overrides.Requests[v1.ResourceMemory])
	assert.EqualValues(t, memoryLimit, overrides.Limits[v1.ResourceMemory])

	// Request equals Limit if not set
	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: memoryLimit,
		},
	}, platformRequirements, assignIfUnset)
	assert.EqualValues(t, memoryLimit, overrides.Requests[v1.ResourceMemory])
	assert.EqualValues(t, memoryLimit, overrides.Limits[v1.ResourceMemory])
}

func TestApplyResourceOverrides_OverrideEphemeralStorage(t *testing.T) {
	ephemeralStorageRequest := resource.MustParse("1")
	overrides := ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageRequest,
		},
	}, v1.ResourceRequirements{}, assignIfUnset)
	assert.EqualValues(t, ephemeralStorageRequest, overrides.Requests[v1.ResourceEphemeralStorage])
	assert.EqualValues(t, ephemeralStorageRequest, overrides.Limits[v1.ResourceEphemeralStorage])

	ephemeralStorageLimit := resource.MustParse("2")
	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageLimit,
		},
	}, v1.ResourceRequirements{}, assignIfUnset)
	assert.EqualValues(t, ephemeralStorageRequest, overrides.Requests[v1.ResourceEphemeralStorage])
	assert.EqualValues(t, ephemeralStorageLimit, overrides.Limits[v1.ResourceEphemeralStorage])

	// Request equals Limit if not set
	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageLimit,
		},
	}, v1.ResourceRequirements{}, assignIfUnset)
	assert.EqualValues(t, ephemeralStorageLimit, overrides.Requests[v1.ResourceEphemeralStorage])
}

func TestApplyResourceOverrides_RemoveStorage(t *testing.T) {
	requestedResourceQuantity := resource.MustParse("1")
	overrides := ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceCPU:              requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
	}, v1.ResourceRequirements{}, assignIfUnset)
	assert.EqualValues(t, v1.ResourceList{
		v1.ResourceMemory:           requestedResourceQuantity,
		v1.ResourceCPU:              requestedResourceQuantity,
		v1.ResourceEphemeralStorage: requestedResourceQuantity,
	}, overrides.Requests)

	assert.EqualValues(t, v1.ResourceList{
		v1.ResourceMemory:           requestedResourceQuantity,
		v1.ResourceCPU:              requestedResourceQuantity,
		v1.ResourceEphemeralStorage: requestedResourceQuantity,
	}, overrides.Limits)
}

func TestApplyResourceOverrides_OverrideGpu(t *testing.T) {
	gpuRequest := resource.MustParse("1")
	overrides := ApplyResourceOverrides(v1.ResourceRequirements{
		Requests: v1.ResourceList{
			ResourceNvidiaGPU: gpuRequest,
		},
	}, v1.ResourceRequirements{}, assignIfUnset)
	assert.EqualValues(t, gpuRequest, overrides.Requests[ResourceNvidiaGPU])

	overrides = ApplyResourceOverrides(v1.ResourceRequirements{
		Limits: v1.ResourceList{
			ResourceNvidiaGPU: gpuRequest,
		},
	}, v1.ResourceRequirements{}, assignIfUnset)
	assert.EqualValues(t, gpuRequest, overrides.Limits[ResourceNvidiaGPU])
}

func TestSanitizeGPUResourceRequirements(t *testing.T) {
	gpuRequest := resource.MustParse("4")
	requirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			resourceGPU: gpuRequest,
		},
	}

	expectedRequirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			ResourceNvidiaGPU: gpuRequest,
		},
	}

	SanitizeGPUResourceRequirements(&requirements)
	assert.EqualValues(t, expectedRequirements, requirements)
}

func TestMergeResources_EmptyIn(t *testing.T) {
	requestedResourceQuantity := resource.MustParse("1")
	expectedResources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceCPU:              requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
	}
	outResources := expectedResources.DeepCopy()
	MergeResources(v1.ResourceRequirements{}, outResources)
	assert.EqualValues(t, *outResources, expectedResources)
}

func TestMergeResources_EmptyOut(t *testing.T) {
	requestedResourceQuantity := resource.MustParse("1")
	expectedResources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceCPU:              requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
	}
	outResources := v1.ResourceRequirements{}
	MergeResources(expectedResources, &outResources)
	assert.EqualValues(t, outResources, expectedResources)
}

func TestMergeResources_PartialRequirements(t *testing.T) {
	requestedResourceQuantity := resource.MustParse("1")
	resourceList := v1.ResourceList{
		v1.ResourceMemory:           requestedResourceQuantity,
		v1.ResourceCPU:              requestedResourceQuantity,
		v1.ResourceEphemeralStorage: requestedResourceQuantity,
	}
	inResources := v1.ResourceRequirements{Requests: resourceList}
	outResources := v1.ResourceRequirements{Limits: resourceList}
	MergeResources(inResources, &outResources)
	assert.EqualValues(t, outResources, v1.ResourceRequirements{
		Requests: resourceList,
		Limits:   resourceList,
	})
}

func TestMergeResources_PartialResourceKeys(t *testing.T) {
	requestedResourceQuantity := resource.MustParse("1")
	resourceList1 := v1.ResourceList{
		v1.ResourceMemory:           requestedResourceQuantity,
		v1.ResourceEphemeralStorage: requestedResourceQuantity,
	}
	resourceList2 := v1.ResourceList{v1.ResourceCPU: requestedResourceQuantity}
	expectedResourceList := v1.ResourceList{
		v1.ResourceCPU:              requestedResourceQuantity,
		v1.ResourceMemory:           requestedResourceQuantity,
		v1.ResourceEphemeralStorage: requestedResourceQuantity,
	}
	inResources := v1.ResourceRequirements{
		Requests: resourceList1,
		Limits:   resourceList2,
	}
	outResources := v1.ResourceRequirements{
		Requests: resourceList2,
		Limits:   resourceList1,
	}
	MergeResources(inResources, &outResources)
	assert.EqualValues(t, outResources, v1.ResourceRequirements{
		Requests: expectedResourceList,
		Limits:   expectedResourceList,
	})
}

func TestToK8sContainer(t *testing.T) {
	taskTemplate := &core.TaskTemplate{
		Type: "test",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: "myimage",
				Args: []string{
					"arg1",
					"arg2",
					"arg3",
				},
				Command: []string{
					"com1",
					"com2",
					"com3",
				},
				Env: []*core.KeyValuePair{
					{
						Key:   "k",
						Value: "v",
					},
				},
			},
		},
	}

	taskReader := &mocks.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(taskTemplate, nil)

	inputReader := &mocks2.InputReader{}
	inputReader.EXPECT().GetInputPath().Return(storage.DataReference("test-data-reference"))
	inputReader.EXPECT().GetInputPrefixPath().Return(storage.DataReference("test-data-reference-prefix"))
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)

	outputWriter := &mocks2.OutputWriter{}
	outputWriter.EXPECT().GetOutputPrefixPath().Return("")
	outputWriter.EXPECT().GetRawOutputPrefix().Return("")
	outputWriter.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputWriter.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskOverrides := mocks.TaskOverrides{}
	mockTaskOverrides.EXPECT().GetResources().Return(&v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: resource.MustParse("1024Mi"),
		},
	})
	mockTaskExecMetadata.EXPECT().GetOverrides().Return(&mockTaskOverrides)
	mockTaskExecutionID := mocks.TaskExecutionID{}
	mockTaskExecutionID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{})
	mockTaskExecutionID.EXPECT().GetGeneratedName().Return("gen_name")
	mockTaskExecMetadata.EXPECT().GetTaskExecutionID().Return(&mockTaskExecutionID)
	mockTaskExecMetadata.EXPECT().GetPlatformResources().Return(&v1.ResourceRequirements{})
	mockTaskExecMetadata.EXPECT().GetEnvironmentVariables().Return(map[string]string{
		"foo": "bar",
	})
	mockTaskExecMetadata.EXPECT().GetNamespace().Return("my-namespace")
	mockTaskExecMetadata.EXPECT().GetConsoleURL().Return("")

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(&mockTaskExecMetadata)
	tCtx.EXPECT().InputReader().Return(inputReader)
	tCtx.EXPECT().TaskReader().Return(taskReader)
	tCtx.EXPECT().OutputWriter().Return(outputWriter)

	cfg := config.GetK8sPluginConfig()
	allow := false
	cfg.DefaultSecurityContext = &v1.SecurityContext{
		AllowPrivilegeEscalation: &allow,
	}
	assert.NoError(t, config.SetK8sPluginConfig(cfg))

	container, err := ToK8sContainer(context.TODO(), tCtx)
	assert.NoError(t, err)
	assert.Equal(t, container.Image, "myimage")
	assert.EqualValues(t, []string{
		"arg1",
		"arg2",
		"arg3",
	}, container.Args)
	assert.EqualValues(t, []string{
		"com1",
		"com2",
		"com3",
	}, container.Command)
	assert.EqualValues(t, []v1.EnvVar{
		{
			Name:  "k",
			Value: "v",
		},
		{
			Name:  "foo",
			Value: "bar",
		},
	}, container.Env)
	errs := validation.IsDNS1123Label(container.Name)
	assert.Nil(t, errs)
	assert.NotNil(t, container.SecurityContext)
	assert.False(t, *container.SecurityContext.AllowPrivilegeEscalation)
}

func getTemplateParametersForTest(resourceRequirements, platformResources *v1.ResourceRequirements, includeConsoleURL bool, consoleURL string) template.Parameters {
	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskExecutionID := mocks.TaskExecutionID{}
	mockTaskExecutionID.EXPECT().GetUniqueNodeID().Return("unique_node_id")
	mockTaskExecutionID.EXPECT().GetGeneratedName().Return("gen_name")
	mockTaskExecutionID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "p1",
			Domain:       "d1",
			Name:         "task_name",
			Version:      "v1",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "node_id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p2",
				Domain:  "d2",
				Name:    "n2",
			},
		},
		RetryAttempt: 1,
	})
	mockTaskExecMetadata.EXPECT().GetTaskExecutionID().Return(&mockTaskExecutionID)

	mockOverrides := mocks.TaskOverrides{}
	mockOverrides.EXPECT().GetResources().Return(resourceRequirements)
	mockTaskExecMetadata.EXPECT().GetOverrides().Return(&mockOverrides)
	mockTaskExecMetadata.EXPECT().GetPlatformResources().Return(platformResources)
	mockTaskExecMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	mockTaskExecMetadata.EXPECT().GetNamespace().Return("my-namespace")
	mockTaskExecMetadata.EXPECT().GetConsoleURL().Return(consoleURL)

	mockInputReader := mocks2.InputReader{}
	mockInputPath := storage.DataReference("s3://input/path")
	mockInputReader.EXPECT().GetInputPath().Return(mockInputPath)
	mockInputReader.EXPECT().GetInputPrefixPath().Return(mockInputPath)
	mockInputReader.On("Get", mock.Anything).Return(nil, nil)

	mockOutputPath := mocks2.OutputFilePaths{}
	mockOutputPathPrefix := storage.DataReference("s3://output/path")
	mockOutputPath.EXPECT().GetRawOutputPrefix().Return(mockOutputPathPrefix)
	mockOutputPath.EXPECT().GetOutputPrefixPath().Return(mockOutputPathPrefix)
	mockOutputPath.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	mockOutputPath.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")

	return template.Parameters{
		TaskExecMetadata:  &mockTaskExecMetadata,
		Inputs:            &mockInputReader,
		OutputPath:        &mockOutputPath,
		IncludeConsoleURL: includeConsoleURL,
	}
}

func TestAddFlyteCustomizationsToContainer(t *testing.T) {
	templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceEphemeralStorage: resource.MustParse("1024Mi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: resource.MustParse("2048Mi"),
		},
	}, nil, false, "")
	container := &v1.Container{
		Command: []string{
			"{{ .Input }}",
		},
		Args: []string{
			"{{ .OutputPrefix }}",
		},
	}
	err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeAssignResources, container)
	assert.NoError(t, err)
	assert.EqualValues(t, container.Args, []string{"s3://output/path"})
	assert.EqualValues(t, container.Command, []string{"s3://input/path"})
	assert.Len(t, container.Resources.Limits, 3)
	assert.Len(t, container.Resources.Requests, 3)
	assert.Len(t, container.Env, 12)
}

func TestAddFlyteCustomizationsToContainer_Resources(t *testing.T) {
	container := &v1.Container{
		Command: []string{
			"{{ .Input }}",
		},
		Args: []string{
			"{{ .OutputPrefix }}",
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1"),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("10"),
			},
		},
	}

	t.Run("merge requests/limits for pod tasks - primary container", func(t *testing.T) {
		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2"),
			},
		}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("20"),
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2"),
			},
		}, false, "")
		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container)
		assert.NoError(t, err)
		assert.True(t, container.Resources.Requests.Cpu().Equal(resource.MustParse("1")))
		assert.True(t, container.Resources.Limits.Cpu().Equal(resource.MustParse("10")))
		assert.True(t, container.Resources.Requests.Memory().Equal(resource.MustParse("2")))
		assert.True(t, container.Resources.Limits.Memory().Equal(resource.MustParse("2")))
	})
	t.Run("enforce merge requests/limits for pod tasks - values from task overrides", func(t *testing.T) {
		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2"),
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("200"),
			},
		}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("2"),
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("20"),
			},
		}, false, "")
		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container)
		assert.NoError(t, err)
		assert.True(t, container.Resources.Requests.Cpu().Equal(resource.MustParse("1")))
		assert.True(t, container.Resources.Limits.Cpu().Equal(resource.MustParse("10")))
		assert.True(t, container.Resources.Requests.Memory().Equal(resource.MustParse("2")))
		assert.True(t, container.Resources.Limits.Memory().Equal(resource.MustParse("20")))
	})
	t.Run("enforce requests/limits for pod tasks - values from container", func(t *testing.T) {
		container := &v1.Container{
			Command: []string{
				"{{ .Input }}",
			},
			Args: []string{
				"{{ .OutputPrefix }}",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("100"),
				},
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("2"),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("10"),
				v1.ResourceMemory: resource.MustParse("20"),
			},
		}, false, "")
		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container)
		assert.NoError(t, err)
		assert.True(t, container.Resources.Requests.Cpu().Equal(resource.MustParse("10")))
		assert.True(t, container.Resources.Limits.Cpu().Equal(resource.MustParse("10")))
		assert.True(t, container.Resources.Requests.Memory().Equal(resource.MustParse("2")))
		assert.True(t, container.Resources.Limits.Memory().Equal(resource.MustParse("2")))
	})
	t.Run("ensure gpu resource overriding works for tasks with pod templates", func(t *testing.T) {
		container := &v1.Container{
			Command: []string{
				"{{ .Input }}",
			},
			Args: []string{
				"{{ .OutputPrefix }}",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					resourceGPU: resource.MustParse("2"), // Tasks with pod templates request resource via the "gpu" key
				},
				Limits: v1.ResourceList{
					resourceGPU: resource.MustParse("2"),
				},
			},
		}

		overrideRequests := v1.ResourceList{
			ResourceNvidiaGPU: resource.MustParse("4"), // Resource overrides specify the "nvidia.com/gpu" key
		}

		overrideLimits := v1.ResourceList{
			ResourceNvidiaGPU: resource.MustParse("4"),
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: overrideRequests,
			Limits:   overrideLimits,
		}, &v1.ResourceRequirements{}, false, "")

		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container)
		assert.NoError(t, err)
		assert.Equal(t, container.Resources.Requests[ResourceNvidiaGPU], overrideRequests[ResourceNvidiaGPU])
		assert.Equal(t, container.Resources.Limits[ResourceNvidiaGPU], overrideLimits[ResourceNvidiaGPU])
	})
}

func TestAddFlyteCustomizationsToContainer_ValidateExistingResources(t *testing.T) {
	container := &v1.Container{
		Command: []string{
			"{{ .Input }}",
		},
		Args: []string{
			"{{ .OutputPrefix }}",
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("100"),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("200"),
			},
		},
	}
	templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("2"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("10"),
			v1.ResourceMemory: resource.MustParse("20"),
		},
	}, false, "")
	err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeEnsureExistingResourcesInRange, container)
	assert.NoError(t, err)

	assert.True(t, container.Resources.Requests.Cpu().Equal(resource.MustParse("10")))
	assert.True(t, container.Resources.Limits.Cpu().Equal(resource.MustParse("10")))
}

func TestAddFlyteCustomizationsToContainer_ValidateEnvFrom(t *testing.T) {
	configMapSource := v1.EnvFromSource{
		ConfigMapRef: &v1.ConfigMapEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "my-configmap",
			},
		},
	}
	secretSource := v1.EnvFromSource{
		SecretRef: &v1.SecretEnvSource{
			LocalObjectReference: v1.LocalObjectReference{
				Name: "my-secret",
			},
		},
	}

	container := &v1.Container{
		Command: []string{
			"{{ .Input }}",
		},
		Args: []string{
			"{{ .OutputPrefix }}",
		},
		EnvFrom: []v1.EnvFromSource{
			configMapSource,
			secretSource,
		},
	}

	err := AddFlyteCustomizationsToContainer(context.TODO(), getTemplateParametersForTest(nil, nil, false, ""), ResourceCustomizationModeEnsureExistingResourcesInRange, container)
	assert.NoError(t, err)

	assert.Len(t, container.EnvFrom, 2)
	assert.Equal(t, container.EnvFrom[0], configMapSource)
	assert.Equal(t, container.EnvFrom[1], secretSource)
}

func TestExtractContainerResourcesFromPodTemplate(t *testing.T) {
	// Define all resource values as named variables for better readability
	primaryCPURequest := resource.MustParse("100m")
	primaryMemoryRequest := resource.MustParse("128Mi")
	primaryCPULimit := resource.MustParse("200m")
	primaryMemoryLimit := resource.MustParse("256Mi")
	otherCPURequest := resource.MustParse("50m")

	expectedResources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    primaryCPURequest,
			v1.ResourceMemory: primaryMemoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    primaryCPULimit,
			v1.ResourceMemory: primaryMemoryLimit,
		},
	}

	otherContainerResources := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: otherCPURequest,
		},
	}

	t.Run("nil pod template returns empty resources", func(t *testing.T) {
		result := ExtractContainerResourcesFromPodTemplate(nil, "test-container", false)
		assert.Equal(t, v1.ResourceRequirements{}, result)
	})

	t.Run("exact container name match - regular containers", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "test-container",
							Resources: expectedResources,
						},
						{
							Name:      "other-container",
							Resources: otherContainerResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "test-container", false)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("exact container name match - init containers", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:      "test-init-container",
							Resources: expectedResources,
						},
						{
							Name:      "other-init-container",
							Resources: otherContainerResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "test-init-container", true)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("fallback to primary container", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "some-container",
							Resources: otherContainerResources,
						},
						{
							Name:      "primary",
							Resources: expectedResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-container", false)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("fallback to primary-init container", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:      "some-init-container",
							Resources: otherContainerResources,
						},
						{
							Name:      "primary-init",
							Resources: expectedResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-init-container", true)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("fallback to default container", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "some-container",
							Resources: otherContainerResources,
						},
						{
							Name:      "default",
							Resources: expectedResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-container", false)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("fallback to default-init container", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:      "some-init-container",
							Resources: otherContainerResources,
						},
						{
							Name:      "default-init",
							Resources: expectedResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-init-container", true)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("no matching container returns empty resources", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "some-container",
							Resources: otherContainerResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-container", false)
		assert.Equal(t, v1.ResourceRequirements{}, result)
	})

	t.Run("priority order - exact match over primary", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "test-container",
							Resources: expectedResources,
						},
						{
							Name:      "primary",
							Resources: otherContainerResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "test-container", false)
		assert.Equal(t, expectedResources, result)
	})

	t.Run("priority order - primary over default", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:      "primary",
							Resources: expectedResources,
						},
						{
							Name:      "default",
							Resources: otherContainerResources,
						},
					},
				},
			},
		}

		result := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-container", false)
		assert.Equal(t, expectedResources, result)
	})
}

func TestAddFlyteCustomizationsToContainerWithPodTemplate(t *testing.T) {
	// Define all resource values as named variables for better readability

	// Container resource values
	containerCPURequest := resource.MustParse("100m")
	containerCPURequestHigh := resource.MustParse("300m")
	containerCPULimitHigh := resource.MustParse("600m")
	containerCPURequestIgnored := resource.MustParse("500m")
	containerCPURequestExceeded := resource.MustParse("1500m")
	containerCPULimitExceeded := resource.MustParse("2000m")

	// Pod template resource values
	podTemplateMemoryRequest := resource.MustParse("256Mi")
	podTemplateMemoryLimit := resource.MustParse("512Mi")

	// Override resource values
	overrideCPURequest := resource.MustParse("200m")

	// Platform resource values
	platformCPURequest := resource.MustParse("100m")
	platformCPULimit := resource.MustParse("1000m")
	platformMemoryRequest := resource.MustParse("128Mi")
	platformMemoryLimit := resource.MustParse("1Gi")

	// GPU resource values
	gpuContainerAmount := resource.MustParse("1")
	gpuPodTemplateAmount := resource.MustParse("2")
	gpuOverrideAmount := resource.MustParse("4")

	t.Run("basic functionality with pod template resources", func(t *testing.T) {
		container := &v1.Container{
			Name: "test-container",
			Command: []string{
				"{{ .Input }}",
			},
			Args: []string{
				"{{ .OutputPrefix }}",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: containerCPURequest,
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryLimit,
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: overrideCPURequest,
			},
		}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    platformCPURequest,
				v1.ResourceMemory: platformMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    platformCPULimit,
				v1.ResourceMemory: platformMemoryLimit,
			},
		}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, podTemplateResources)
		assert.NoError(t, err)

		// Check that template rendering worked
		assert.Equal(t, []string{"s3://input/path"}, container.Command)
		assert.Equal(t, []string{"s3://output/path"}, container.Args)

		// Check that resources were properly merged
		// Priority: overrideResources > container.Resources > podTemplateResources > platformResources
		assert.True(t, container.Resources.Requests.Cpu().Equal(overrideCPURequest))          // from override
		assert.True(t, container.Resources.Requests.Memory().Equal(podTemplateMemoryRequest)) // from pod template
		assert.True(t, container.Resources.Limits.Memory().Equal(podTemplateMemoryLimit))     // from pod template

		// Check that environment variables were decorated
		assert.Greater(t, len(container.Env), 0)
	})

	t.Run("resource priority order - AssignResources mode", func(t *testing.T) {
		container := &v1.Container{
			Name: "test-container",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: containerCPURequestIgnored, // This should be ignored in AssignResources mode
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryRequest,
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: overrideCPURequest,
			},
		}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    platformCPURequest,
				v1.ResourceMemory: platformMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    platformCPULimit,
				v1.ResourceMemory: platformMemoryLimit,
			},
		}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeAssignResources, container, podTemplateResources)
		assert.NoError(t, err)

		// In AssignResources mode, container.Resources should be ignored, only overrideResources and platformResources matter
		assert.True(t, container.Resources.Requests.Cpu().Equal(overrideCPURequest)) // from override
		assert.True(t, container.Resources.Limits.Cpu().Equal(overrideCPURequest))   // assigned from request
	})

	t.Run("resource priority order - MergeExistingResources mode", func(t *testing.T) {
		container := &v1.Container{
			Name: "test-container",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: containerCPURequestHigh,
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: containerCPULimitHigh,
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryLimit,
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: overrideCPURequest, // This should override container CPU
			},
		}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    platformCPURequest,
				v1.ResourceMemory: platformMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    platformCPULimit,
				v1.ResourceMemory: platformMemoryLimit,
			},
		}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, podTemplateResources)
		assert.NoError(t, err)

		// Priority: overrideResources > container.Resources > podTemplateResources > platformResources
		assert.True(t, container.Resources.Requests.Cpu().Equal(overrideCPURequest)) // from override
		// Since override only has CPU request, ApplyResourceOverrides sets the limit equal to the request
		assert.True(t, container.Resources.Limits.Cpu().Equal(overrideCPURequest))            // set equal to request since override had no limit
		assert.True(t, container.Resources.Requests.Memory().Equal(podTemplateMemoryRequest)) // from pod template
		assert.True(t, container.Resources.Limits.Memory().Equal(podTemplateMemoryLimit))     // from pod template
	})

	t.Run("resource priority order - EnsureExistingResourcesInRange mode", func(t *testing.T) {
		container := &v1.Container{
			Name: "test-container",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: containerCPURequestExceeded, // This exceeds platform limit
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: containerCPULimitExceeded, // This exceeds platform limit
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryLimit,
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    platformCPURequest,
				v1.ResourceMemory: platformMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    platformCPULimit, // Platform limit
				v1.ResourceMemory: platformMemoryLimit,
			},
		}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeEnsureExistingResourcesInRange, container, podTemplateResources)
		assert.NoError(t, err)

		// Resources should be capped at platform limits
		assert.True(t, container.Resources.Requests.Cpu().Equal(platformCPULimit))            // capped to platform limit
		assert.True(t, container.Resources.Limits.Cpu().Equal(platformCPULimit))              // capped to platform limit
		assert.True(t, container.Resources.Requests.Memory().Equal(podTemplateMemoryRequest)) // from pod template
		assert.True(t, container.Resources.Limits.Memory().Equal(podTemplateMemoryLimit))     // from pod template
	})

	t.Run("nil pod template resources", func(t *testing.T) {
		container := &v1.Container{
			Name: "test-container",
			Command: []string{
				"echo",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: containerCPURequest,
				},
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    platformCPURequest,
				v1.ResourceMemory: platformMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    platformCPULimit,
				v1.ResourceMemory: platformMemoryLimit,
			},
		}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, nil)
		assert.NoError(t, err)

		// Should still work with nil pod template resources
		assert.True(t, container.Resources.Requests.Cpu().Equal(containerCPURequest))
		// Memory gets platform limit (1Gi) instead of platform request (128Mi) due to resource adjustment logic
		assert.True(t, container.Resources.Requests.Memory().Equal(platformMemoryLimit)) // from platform limit due to resource adjustment
	})

	t.Run("template rendering with console URL", func(t *testing.T) {
		consoleBaseURL := "https://flyte.example.com/console"
		expectedConsoleURL := "https://flyte.example.com/console/projects/p2/domains/d2/executions/n2/nodeId/unique_node_id/nodes"

		container := &v1.Container{
			Name: "test-container",
			Command: []string{
				"{{ .Input }}",
			},
			Args: []string{
				"{{ .OutputPrefix }}",
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    platformCPURequest,
				v1.ResourceMemory: platformMemoryRequest,
			},
		}, true, consoleBaseURL)

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeAssignResources, container, nil)
		assert.NoError(t, err)

		// Check that template rendering worked
		assert.Equal(t, []string{"s3://input/path"}, container.Command)
		assert.Equal(t, []string{"s3://output/path"}, container.Args)

		// Check that console URL was added to environment variables
		consoleURLFound := false
		for _, env := range container.Env {
			if env.Name == "FLYTE_EXECUTION_URL" && env.Value == expectedConsoleURL {
				consoleURLFound = true
				break
			}
		}
		assert.True(t, consoleURLFound, "Console URL should be present in environment variables with correct format")
	})

	t.Run("GPU resource handling with pod template", func(t *testing.T) {
		container := &v1.Container{
			Name: "test-container",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					resourceGPU: gpuContainerAmount, // Tasks with pod templates use "gpu" key
				},
				Limits: v1.ResourceList{
					resourceGPU: gpuContainerAmount,
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				ResourceNvidiaGPU: gpuPodTemplateAmount, // Pod template uses nvidia.com/gpu key
			},
			Limits: v1.ResourceList{
				ResourceNvidiaGPU: gpuPodTemplateAmount,
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				ResourceNvidiaGPU: gpuOverrideAmount, // Override uses nvidia.com/gpu key
			},
			Limits: v1.ResourceList{
				ResourceNvidiaGPU: gpuOverrideAmount,
			},
		}, &v1.ResourceRequirements{}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, podTemplateResources)
		assert.NoError(t, err)

		// In MergeExistingResources mode:
		// 1) overrideResources (4 nvidia.com/gpu) are merged first
		// 2) Then pod template resources are applied for missing resources
		// 3) The original container "gpu" resources are converted by SanitizeGPUResourceRequirements function
		// But the function doesn't automatically convert, so the final result uses pod template values
		assert.True(t, container.Resources.Requests[ResourceNvidiaGPU].Equal(gpuPodTemplateAmount)) // from pod template
		assert.True(t, container.Resources.Limits[ResourceNvidiaGPU].Equal(gpuPodTemplateAmount))   // from pod template
		// Check that CPU and memory get platform defaults
		assert.True(t, container.Resources.Requests.Cpu().Equal(platformCPULimit)) // platform default
		assert.True(t, container.Resources.Limits.Cpu().Equal(platformCPULimit))   // platform default
	})
}
