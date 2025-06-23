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

func TestAddFlyteCustomizationsToContainerWithPodTemplate(t *testing.T) {
	t.Run("merge pod template resources with proper priority", func(t *testing.T) {
		container := &v1.Container{
			Command: []string{
				"{{ .Input }}",
			},
			Args: []string{
				"{{ .OutputPrefix }}",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:   resource.MustParse("1"), // Container inline resource (priority 2)
					"nvidia.com/gpu": resource.MustParse("2"), // Container inline resource (priority 2)
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:   resource.MustParse("10"),
					"nvidia.com/gpu": resource.MustParse("2"),
				},
			},
		}

		// Pod template resources (priority 3)
		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("5"),   // Should NOT override container CPU
				v1.ResourceMemory: resource.MustParse("8Gi"), // Should be used since not in container
				"rdma/infiniband": resource.MustParse("4"),   // Should be used since not in container
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("50"),
				v1.ResourceMemory: resource.MustParse("16Gi"),
				"rdma/infiniband": resource.MustParse("4"),
			},
		}

		// Override resources (priority 1)
		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("16Gi"), // Should override pod template memory
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("32Gi"),
			},
		}, &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("2"),
				v1.ResourceMemory: resource.MustParse("4Gi"),
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("20"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
			},
		}, false, "")

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, podTemplateResources)
		assert.NoError(t, err)

		// Verify resource priority:
		// 1. Override resources should take precedence but are subject to platform limits
		assert.True(t, container.Resources.Requests.Memory().Equal(resource.MustParse("8Gi")), "Memory request should be capped at platform limit")
		assert.True(t, container.Resources.Limits.Memory().Equal(resource.MustParse("8Gi")), "Memory limit should be capped at platform limit")

		// 2. Container inline resources should be preserved when not overridden
		assert.True(t, container.Resources.Requests.Cpu().Equal(resource.MustParse("1")), "Container CPU request should be preserved")
		assert.True(t, container.Resources.Requests["nvidia.com/gpu"].Equal(resource.MustParse("2")), "Container GPU request should be preserved")

		// 3. Pod template resources should be used when not present in container or overrides
		assert.True(t, container.Resources.Requests["rdma/infiniband"].Equal(resource.MustParse("4")), "Pod template RDMA should be used")
		assert.True(t, container.Resources.Limits["rdma/infiniband"].Equal(resource.MustParse("4")), "Pod template RDMA limit should be used")
	})
}

func TestExtractContainerResourcesFromPodTemplate(t *testing.T) {
	t.Run("extract resources from pod template with exact name match", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "my-container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("55"),
									v1.ResourceMemory: resource.MustParse("1837Gi"),
									"nvidia.com/gpu":  resource.MustParse("120"),
									"rdma/infiniband": resource.MustParse("63"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("55"),
									v1.ResourceMemory: resource.MustParse("1837Gi"),
									"nvidia.com/gpu":  resource.MustParse("120"),
									"rdma/infiniband": resource.MustParse("63"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "my-container", false)

		assert.True(t, resources.Requests.Cpu().Equal(resource.MustParse("55")))
		assert.True(t, resources.Requests.Memory().Equal(resource.MustParse("1837Gi")))
		assert.True(t, resources.Requests["nvidia.com/gpu"].Equal(resource.MustParse("120")))
		assert.True(t, resources.Requests["rdma/infiniband"].Equal(resource.MustParse("63")))
	})

	t.Run("extract resources from pod template with primary fallback", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "primary",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent", false)

		assert.True(t, resources.Requests.Cpu().Equal(resource.MustParse("2")))
	})

	t.Run("extract resources from pod template with default fallback", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "default",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("4Gi"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent", false)

		assert.True(t, resources.Requests.Memory().Equal(resource.MustParse("4Gi")))
	})

	t.Run("return empty resources when no match found", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "some-other-container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent", false)

		assert.Empty(t, resources.Requests)
		assert.Empty(t, resources.Limits)
	})

	t.Run("extract resources from pod template with exact init container name match", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "my-init-container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("100m"),
									v1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("200m"),
									v1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "my-init-container", true)

		assert.True(t, resources.Requests.Cpu().Equal(resource.MustParse("100m")))
		assert.True(t, resources.Requests.Memory().Equal(resource.MustParse("256Mi")))
		assert.True(t, resources.Limits.Cpu().Equal(resource.MustParse("200m")))
		assert.True(t, resources.Limits.Memory().Equal(resource.MustParse("512Mi")))
	})

	t.Run("extract resources from pod template with primary-init fallback", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "primary-init",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("500m"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-init", true)

		assert.True(t, resources.Requests.Cpu().Equal(resource.MustParse("500m")))
	})

	t.Run("extract resources from pod template with default-init fallback", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "default-init",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-init", true)

		assert.True(t, resources.Requests.Memory().Equal(resource.MustParse("1Gi")))
	})

	t.Run("prefer exact match over fallback templates", func(t *testing.T) {
		podTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "specific-container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("100m"),
								},
							},
						},
						{
							Name: "primary",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("200m"),
								},
							},
						},
					},
					InitContainers: []v1.Container{
						{
							Name: "specific-init-container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("50m"),
								},
							},
						},
						{
							Name: "primary-init",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("150m"),
								},
							},
						},
					},
				},
			},
		}

		// Test exact match for regular container
		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "specific-container", false)
		assert.True(t, resources.Requests.Cpu().Equal(resource.MustParse("100m")))

		// Test exact match for init container
		resources = ExtractContainerResourcesFromPodTemplate(podTemplate, "specific-init-container", true)
		assert.True(t, resources.Requests.Cpu().Equal(resource.MustParse("50m")))
	})
}
