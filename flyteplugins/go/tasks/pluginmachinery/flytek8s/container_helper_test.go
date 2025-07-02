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
			resource.MustParse("10"), resource.MustParse("20"), false)
		assert.True(t, res.Request.Equal(resource.MustParse("1")))
		assert.True(t, res.Limit.Equal(resource.MustParse("2")))
	})
	t.Run("Assign unset Request from Limit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			zeroQuantity, resource.MustParse("2"),
			resource.MustParse("10"), resource.MustParse("20"), false)
		assert.True(t, res.Request.Equal(resource.MustParse("2")))
		assert.True(t, res.Limit.Equal(resource.MustParse("2")))
	})
	t.Run("Assign unset Limit from Request", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("2"), zeroQuantity,
			resource.MustParse("10"), resource.MustParse("20"), false)
		assert.Equal(t, resource.MustParse("2"), res.Request)
		assert.Equal(t, resource.MustParse("2"), res.Limit)
	})
	t.Run("Assign from platform defaults", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			zeroQuantity, zeroQuantity,
			resource.MustParse("10"), resource.MustParse("20"), false)
		assert.Equal(t, resource.MustParse("10"), res.Request)
		assert.Equal(t, resource.MustParse("10"), res.Limit)
	})
	t.Run("Adjust Limit when Request > Limit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("10"), resource.MustParse("2"),
			resource.MustParse("10"), resource.MustParse("20"), false)
		assert.Equal(t, resource.MustParse("2"), res.Request)
		assert.Equal(t, resource.MustParse("2"), res.Limit)
	})
	t.Run("Adjust Limit > platformLimit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("1"), resource.MustParse("40"),
			resource.MustParse("10"), resource.MustParse("20"), false)
		assert.True(t, res.Request.Equal(resource.MustParse("1")))
		assert.True(t, res.Limit.Equal(resource.MustParse("20")))
	})
	t.Run("Adjust Request, Limit > platformLimit", func(t *testing.T) {
		res := AdjustOrDefaultResource(
			resource.MustParse("40"), resource.MustParse("50"),
			resource.MustParse("10"), resource.MustParse("20"), false)
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

func TestMergeResourcesIfMissing(t *testing.T) {
	// Define test resource values
	sourceCPURequest := resource.MustParse("100m")
	sourceCPULimit := resource.MustParse("200m")
	sourceMemoryRequest := resource.MustParse("128Mi")
	sourceMemoryLimit := resource.MustParse("256Mi")
	sourceStorageRequest := resource.MustParse("1Gi")

	destinationCPURequest := resource.MustParse("150m")
	destinationCPULimit := resource.MustParse("300m")
	destinationMemoryRequest := resource.MustParse("256Mi")

	t.Run("merge resources when destination is empty", func(t *testing.T) {
		sourceResources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    sourceCPURequest,
				v1.ResourceMemory: sourceMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    sourceCPULimit,
				v1.ResourceMemory: sourceMemoryLimit,
			},
		}

		destinationResources := v1.ResourceRequirements{}

		MergeResourcesIfMissing(sourceResources, &destinationResources)

		assert.Equal(t, sourceCPURequest, destinationResources.Requests[v1.ResourceCPU])
		assert.Equal(t, sourceMemoryRequest, destinationResources.Requests[v1.ResourceMemory])
		assert.Equal(t, sourceCPULimit, destinationResources.Limits[v1.ResourceCPU])
		assert.Equal(t, sourceMemoryLimit, destinationResources.Limits[v1.ResourceMemory])
	})

	t.Run("preserve existing resources and only add missing ones", func(t *testing.T) {
		sourceResources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:              sourceCPURequest,
				v1.ResourceMemory:           sourceMemoryRequest,
				v1.ResourceEphemeralStorage: sourceStorageRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    sourceCPULimit,
				v1.ResourceMemory: sourceMemoryLimit,
			},
		}

		destinationResources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    destinationCPURequest,    // This should NOT be overwritten
				v1.ResourceMemory: destinationMemoryRequest, // This should NOT be overwritten
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU: destinationCPULimit, // This should NOT be overwritten
			},
		}

		MergeResourcesIfMissing(sourceResources, &destinationResources)

		// Existing resources should be preserved
		assert.Equal(t, destinationCPURequest, destinationResources.Requests[v1.ResourceCPU])
		assert.Equal(t, destinationMemoryRequest, destinationResources.Requests[v1.ResourceMemory])
		assert.Equal(t, destinationCPULimit, destinationResources.Limits[v1.ResourceCPU])

		// Missing resources should be added
		assert.Equal(t, sourceStorageRequest, destinationResources.Requests[v1.ResourceEphemeralStorage])
		assert.Equal(t, sourceMemoryLimit, destinationResources.Limits[v1.ResourceMemory])
	})

	t.Run("handle nil source resources", func(t *testing.T) {
		sourceResources := v1.ResourceRequirements{}

		destinationResources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: destinationCPURequest,
			},
		}

		originalDestination := destinationResources.DeepCopy()
		MergeResourcesIfMissing(sourceResources, &destinationResources)

		// Destination should remain unchanged
		assert.Equal(t, *originalDestination, destinationResources)
	})

	t.Run("handle partial source resources", func(t *testing.T) {
		sourceResources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: sourceMemoryRequest,
			},
			// No limits
		}

		destinationResources := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: destinationCPURequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU: destinationCPULimit,
			},
		}

		MergeResourcesIfMissing(sourceResources, &destinationResources)

		// Existing resources should be preserved
		assert.Equal(t, destinationCPURequest, destinationResources.Requests[v1.ResourceCPU])
		assert.Equal(t, destinationCPULimit, destinationResources.Limits[v1.ResourceCPU])

		// Missing resources should be added
		assert.Equal(t, sourceMemoryRequest, destinationResources.Requests[v1.ResourceMemory])
	})
}

func TestExtractContainerResourcesFromPodTemplate(t *testing.T) {
	// Define test resource values
	primaryCPURequest := resource.MustParse("100m")
	primaryCPULimit := resource.MustParse("200m")
	primaryMemoryRequest := resource.MustParse("128Mi")
	primaryMemoryLimit := resource.MustParse("256Mi")

	defaultCPURequest := resource.MustParse("50m")
	defaultCPULimit := resource.MustParse("100m")
	defaultMemoryRequest := resource.MustParse("64Mi")
	defaultMemoryLimit := resource.MustParse("128Mi")

	namedCPURequest := resource.MustParse("150m")
	namedCPULimit := resource.MustParse("300m")
	namedMemoryRequest := resource.MustParse("192Mi")
	namedMemoryLimit := resource.MustParse("384Mi")

	initPrimaryCPURequest := resource.MustParse("25m")
	initPrimaryCPULimit := resource.MustParse("50m")
	initPrimaryMemoryRequest := resource.MustParse("32Mi")
	initPrimaryMemoryLimit := resource.MustParse("64Mi")

	initDefaultCPURequest := resource.MustParse("10m")
	initDefaultCPULimit := resource.MustParse("20m")
	initDefaultMemoryRequest := resource.MustParse("16Mi")
	initDefaultMemoryLimit := resource.MustParse("32Mi")

	podTemplate := &v1.PodTemplate{
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "primary",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    primaryCPURequest,
								v1.ResourceMemory: primaryMemoryRequest,
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    primaryCPULimit,
								v1.ResourceMemory: primaryMemoryLimit,
							},
						},
					},
					{
						Name: "default",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    defaultCPURequest,
								v1.ResourceMemory: defaultMemoryRequest,
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    defaultCPULimit,
								v1.ResourceMemory: defaultMemoryLimit,
							},
						},
					},
					{
						Name: "named-container",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    namedCPURequest,
								v1.ResourceMemory: namedMemoryRequest,
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    namedCPULimit,
								v1.ResourceMemory: namedMemoryLimit,
							},
						},
					},
				},
				InitContainers: []v1.Container{
					{
						Name: "primary-init",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    initPrimaryCPURequest,
								v1.ResourceMemory: initPrimaryMemoryRequest,
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    initPrimaryCPULimit,
								v1.ResourceMemory: initPrimaryMemoryLimit,
							},
						},
					},
					{
						Name: "default-init",
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    initDefaultCPURequest,
								v1.ResourceMemory: initDefaultMemoryRequest,
							},
							Limits: v1.ResourceList{
								v1.ResourceCPU:    initDefaultCPULimit,
								v1.ResourceMemory: initDefaultMemoryLimit,
							},
						},
					},
				},
			},
		},
	}

	t.Run("extract resources from exact container name match", func(t *testing.T) {
		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "named-container", false)

		assert.Equal(t, namedCPURequest, resources.Requests[v1.ResourceCPU])
		assert.Equal(t, namedMemoryRequest, resources.Requests[v1.ResourceMemory])
		assert.Equal(t, namedCPULimit, resources.Limits[v1.ResourceCPU])
		assert.Equal(t, namedMemoryLimit, resources.Limits[v1.ResourceMemory])
	})

	t.Run("extract resources from primary container when no exact match", func(t *testing.T) {
		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-container", false)

		assert.Equal(t, primaryCPURequest, resources.Requests[v1.ResourceCPU])
		assert.Equal(t, primaryMemoryRequest, resources.Requests[v1.ResourceMemory])
		assert.Equal(t, primaryCPULimit, resources.Limits[v1.ResourceCPU])
		assert.Equal(t, primaryMemoryLimit, resources.Limits[v1.ResourceMemory])
	})

	t.Run("extract resources from default container when no primary", func(t *testing.T) {
		podTemplateWithoutPrimary := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "default",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    defaultCPURequest,
									v1.ResourceMemory: defaultMemoryRequest,
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    defaultCPULimit,
									v1.ResourceMemory: defaultMemoryLimit,
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplateWithoutPrimary, "non-existent-container", false)

		assert.Equal(t, defaultCPURequest, resources.Requests[v1.ResourceCPU])
		assert.Equal(t, defaultMemoryRequest, resources.Requests[v1.ResourceMemory])
		assert.Equal(t, defaultCPULimit, resources.Limits[v1.ResourceCPU])
		assert.Equal(t, defaultMemoryLimit, resources.Limits[v1.ResourceMemory])
	})

	t.Run("extract resources from init containers with primary-init", func(t *testing.T) {
		resources := ExtractContainerResourcesFromPodTemplate(podTemplate, "non-existent-init", true)

		assert.Equal(t, initPrimaryCPURequest, resources.Requests[v1.ResourceCPU])
		assert.Equal(t, initPrimaryMemoryRequest, resources.Requests[v1.ResourceMemory])
		assert.Equal(t, initPrimaryCPULimit, resources.Limits[v1.ResourceCPU])
		assert.Equal(t, initPrimaryMemoryLimit, resources.Limits[v1.ResourceMemory])
	})

	t.Run("extract resources from init containers with default-init", func(t *testing.T) {
		podTemplateWithoutPrimaryInit := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "default-init",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    initDefaultCPURequest,
									v1.ResourceMemory: initDefaultMemoryRequest,
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    initDefaultCPULimit,
									v1.ResourceMemory: initDefaultMemoryLimit,
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(podTemplateWithoutPrimaryInit, "non-existent-init", true)

		assert.Equal(t, initDefaultCPURequest, resources.Requests[v1.ResourceCPU])
		assert.Equal(t, initDefaultMemoryRequest, resources.Requests[v1.ResourceMemory])
		assert.Equal(t, initDefaultCPULimit, resources.Limits[v1.ResourceCPU])
		assert.Equal(t, initDefaultMemoryLimit, resources.Limits[v1.ResourceMemory])
	})

	t.Run("return empty resources when no match found", func(t *testing.T) {
		emptyPodTemplate := &v1.PodTemplate{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "some-other-container",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU: resource.MustParse("1m"),
								},
							},
						},
					},
				},
			},
		}

		resources := ExtractContainerResourcesFromPodTemplate(emptyPodTemplate, "non-existent-container", false)

		assert.Empty(t, resources.Requests)
		assert.Empty(t, resources.Limits)
	})

	t.Run("return empty resources when pod template is nil", func(t *testing.T) {
		resources := ExtractContainerResourcesFromPodTemplate(nil, "any-container", false)

		assert.Empty(t, resources.Requests)
		assert.Empty(t, resources.Limits)
	})
}

func TestAddFlyteCustomizationsToContainerWithPodTemplate(t *testing.T) {
	// Define test resource values
	containerCPURequest := resource.MustParse("100m")
	containerCPULimit := resource.MustParse("200m")
	containerMemoryRequest := resource.MustParse("128Mi")
	containerMemoryLimit := resource.MustParse("256Mi")

	podTemplateCPURequest := resource.MustParse("50m")
	podTemplateCPULimit := resource.MustParse("150m")
	podTemplateMemoryRequest := resource.MustParse("64Mi")
	podTemplateMemoryLimit := resource.MustParse("192Mi")
	podTemplateStorageRequest := resource.MustParse("1Gi")
	podTemplateStorageLimit := resource.MustParse("2Gi")

	overrideCPURequest := resource.MustParse("300m")
	overrideCPULimit := resource.MustParse("400m")
	overrideMemoryRequest := resource.MustParse("512Mi")
	overrideMemoryLimit := resource.MustParse("1Gi")

	platformCPURequest := resource.MustParse("10m")
	platformCPULimit := resource.MustParse("500m")
	platformMemoryRequest := resource.MustParse("32Mi")
	platformMemoryLimit := resource.MustParse("2Gi")

	ExceedsCPURequest := resource.MustParse("1000m") // Exceeds platform limit
	ExceedsCPULimit := resource.MustParse("2000m")   // Exceeds platform limit

	t.Run("merge existing resources mode with pod template", func(t *testing.T) {
		container := &v1.Container{
			Name:    "test-container",
			Command: []string{"{{ .Input }}"},
			Args:    []string{"{{ .OutputPrefix }}"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    containerCPURequest,
					v1.ResourceMemory: containerMemoryRequest,
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    containerCPULimit,
					v1.ResourceMemory: containerMemoryLimit,
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:              podTemplateCPURequest,     // Should not override existing
				v1.ResourceMemory:           podTemplateMemoryRequest,  // Should not override existing
				v1.ResourceEphemeralStorage: podTemplateStorageRequest, // Should be added as missing
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:              podTemplateCPULimit,     // Should not override existing
				v1.ResourceMemory:           podTemplateMemoryLimit,  // Should not override existing
				v1.ResourceEphemeralStorage: podTemplateStorageLimit, // Should be added as missing
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

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, podTemplateResources)
		assert.NoError(t, err)

		// Container resources should be preserved (higher priority)
		assert.Equal(t, containerCPURequest, container.Resources.Requests[v1.ResourceCPU])
		assert.Equal(t, containerMemoryRequest, container.Resources.Requests[v1.ResourceMemory])
		assert.Equal(t, containerCPULimit, container.Resources.Limits[v1.ResourceCPU])
		assert.Equal(t, containerMemoryLimit, container.Resources.Limits[v1.ResourceMemory])

		// Pod template resources should be added for missing resources
		assert.Equal(t, podTemplateStorageRequest, container.Resources.Requests[v1.ResourceEphemeralStorage])
		assert.Equal(t, podTemplateStorageLimit, container.Resources.Limits[v1.ResourceEphemeralStorage])
	})

	t.Run("merge existing resources mode with overrides and pod template", func(t *testing.T) {
		container := &v1.Container{
			Name:    "test-container",
			Command: []string{"{{ .Input }}"},
			Args:    []string{"{{ .OutputPrefix }}"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceMemory: containerMemoryRequest, // Should be overridden by task override
				},
				Limits: v1.ResourceList{
					v1.ResourceMemory: containerMemoryLimit, // Should be overridden by task override
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:              podTemplateCPURequest,     // Should be added as missing
				v1.ResourceEphemeralStorage: podTemplateStorageRequest, // Should be added as missing
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:              podTemplateCPULimit,     // Should be added as missing
				v1.ResourceEphemeralStorage: podTemplateStorageLimit, // Should be added as missing
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: overrideMemoryRequest, // Should override container memory
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: overrideMemoryLimit, // Should override container memory
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

		// Override resources should take precedence (highest priority)
		assert.Equal(t, overrideMemoryRequest, container.Resources.Requests[v1.ResourceMemory])
		assert.Equal(t, overrideMemoryLimit, container.Resources.Limits[v1.ResourceMemory])

		// Pod template resources should be added for missing resources
		assert.Equal(t, podTemplateCPURequest, container.Resources.Requests[v1.ResourceCPU])
		assert.Equal(t, podTemplateStorageRequest, container.Resources.Requests[v1.ResourceEphemeralStorage])
		assert.Equal(t, podTemplateCPULimit, container.Resources.Limits[v1.ResourceCPU])
		assert.Equal(t, podTemplateStorageLimit, container.Resources.Limits[v1.ResourceEphemeralStorage])
	})

	t.Run("ensure existing resources in range mode with pod template", func(t *testing.T) {
		container := &v1.Container{
			Name:    "test-container",
			Command: []string{"{{ .Input }}"},
			Args:    []string{"{{ .OutputPrefix }}"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: ExceedsCPURequest, // Should be limited by platform
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: ExceedsCPULimit, // Should be limited by platform
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryRequest, // Should be added as missing
			},
			Limits: v1.ResourceList{
				v1.ResourceMemory: podTemplateMemoryLimit, // Should be added as missing
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

		err := AddFlyteCustomizationsToContainerWithPodTemplate(context.TODO(), templateParameters, ResourceCustomizationModeEnsureExistingResourcesInRange, container, podTemplateResources)
		assert.NoError(t, err)

		// Container CPU resources should be limited by platform limits
		assert.Equal(t, platformCPULimit, container.Resources.Requests[v1.ResourceCPU])
		assert.Equal(t, platformCPULimit, container.Resources.Limits[v1.ResourceCPU])

		// Pod template resources should be added for missing resources
		assert.Equal(t, podTemplateMemoryRequest, container.Resources.Requests[v1.ResourceMemory])
		assert.Equal(t, podTemplateMemoryLimit, container.Resources.Limits[v1.ResourceMemory])
	})

	t.Run("assign resources mode ignores pod template", func(t *testing.T) {
		container := &v1.Container{
			Name:    "test-container",
			Command: []string{"{{ .Input }}"},
			Args:    []string{"{{ .OutputPrefix }}"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: containerCPURequest, // Should be ignored in assign mode
				},
			},
		}

		podTemplateResources := &v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    podTemplateCPURequest,    // Should be ignored in assign mode
				v1.ResourceMemory: podTemplateMemoryRequest, // Should be ignored in assign mode
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    overrideCPURequest,
				v1.ResourceMemory: overrideMemoryRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceCPU:    overrideCPULimit,
				v1.ResourceMemory: overrideMemoryLimit,
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

		// Should use override resources and ignore both container and pod template resources
		assert.Equal(t, overrideCPURequest, container.Resources.Requests[v1.ResourceCPU])
		assert.Equal(t, overrideMemoryRequest, container.Resources.Requests[v1.ResourceMemory])
		assert.Equal(t, overrideCPULimit, container.Resources.Limits[v1.ResourceCPU])
		assert.Equal(t, overrideMemoryLimit, container.Resources.Limits[v1.ResourceMemory])
	})

	t.Run("handle nil pod template resources", func(t *testing.T) {
		container := &v1.Container{
			Name:    "test-container",
			Command: []string{"{{ .Input }}"},
			Args:    []string{"{{ .OutputPrefix }}"},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    containerCPURequest,
					v1.ResourceMemory: containerMemoryRequest,
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

		// Should still work with container resources
		assert.Equal(t, containerCPURequest, container.Resources.Requests[v1.ResourceCPU])
		assert.Equal(t, containerMemoryRequest, container.Resources.Requests[v1.ResourceMemory])
		assert.Equal(t, containerCPURequest, container.Resources.Limits[v1.ResourceCPU])       // Should be set to request value
		assert.Equal(t, containerMemoryRequest, container.Resources.Limits[v1.ResourceMemory]) // Should be set to request value
	})
}

func TestApplyResourceOverrides_OverrideCpuAllowFloat(t *testing.T) {
	cfg := config.GetK8sPluginConfig()
	cfg.AllowCPULimitToFloatFromRequest = true
	err := config.SetK8sPluginConfig(cfg)
	assert.NoError(t, err)
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
	_, ok := overrides.Limits[v1.ResourceCPU]
	assert.False(t, ok)
}

func TestApplyResourceOverrides_EmptyCpuLimitAllowFloat(t *testing.T) {
	cfg := config.GetK8sPluginConfig()
	cfg.AllowCPULimitToFloatFromRequest = true
	err := config.SetK8sPluginConfig(cfg)
	assert.NoError(t, err)

	{
		platformRequirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("3"),
			},
		}
		cpuRequest := resource.MustParse("1")
		//tenCpuRequest := resource.MustParse("10")
		overrides := ApplyResourceOverrides(v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: cpuRequest,
			},
		}, platformRequirements, assignIfUnset)
		assert.EqualValues(t, cpuRequest, overrides.Requests[v1.ResourceCPU])
		_, ok := overrides.Limits[v1.ResourceCPU]
		assert.False(t, ok)
	}
	{
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
		_, ok := overrides.Limits[v1.ResourceCPU]
		assert.False(t, ok)
	}
}
