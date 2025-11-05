package flytek8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	mocks2 "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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
	t.Run("nil accelerator defaults to NVIDIA GPU", func(t *testing.T) {
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

		SanitizeGPUResourceRequirements(&requirements, nil)
		assert.EqualValues(t, expectedRequirements, requirements)
	})

	t.Run("NVIDIA_GPU device class", func(t *testing.T) {
		gpuRequest := resource.MustParse("2")
		requirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				resourceGPU: gpuRequest,
			},
			Limits: v1.ResourceList{
				resourceGPU: gpuRequest,
			},
		}

		accelerator := &core.GPUAccelerator{
			Device:      "nvidia-tesla-a100",
			DeviceClass: core.GPUAccelerator_NVIDIA_GPU,
		}

		expectedRequirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceName("nvidia.com/gpu"): gpuRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceName("nvidia.com/gpu"): gpuRequest,
			},
		}

		SanitizeGPUResourceRequirements(&requirements, accelerator)
		assert.EqualValues(t, expectedRequirements, requirements)
	})

	t.Run("GOOGLE_TPU device class", func(t *testing.T) {
		tpuRequest := resource.MustParse("4")
		requirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				resourceGPU: tpuRequest,
			},
			Limits: v1.ResourceList{
				resourceGPU: tpuRequest,
			},
		}

		accelerator := &core.GPUAccelerator{
			Device:      "tpu-v4",
			DeviceClass: core.GPUAccelerator_GOOGLE_TPU,
		}

		expectedRequirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceName("google.com/tpu"): tpuRequest,
			},
			Limits: v1.ResourceList{
				v1.ResourceName("google.com/tpu"): tpuRequest,
			},
		}

		SanitizeGPUResourceRequirements(&requirements, accelerator)
		assert.EqualValues(t, expectedRequirements, requirements)
	})

	t.Run("AMAZON_NEURON device class", func(t *testing.T) {
		neuronRequest := resource.MustParse("1")
		requirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				resourceGPU: neuronRequest,
			},
		}

		accelerator := &core.GPUAccelerator{
			Device:      "inferentia2",
			DeviceClass: core.GPUAccelerator_AMAZON_NEURON,
		}

		expectedRequirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceName("aws.amazon.com/neuron"): neuronRequest,
			},
		}

		SanitizeGPUResourceRequirements(&requirements, accelerator)
		assert.EqualValues(t, expectedRequirements, requirements)
	})

	t.Run("AMD_GPU device class", func(t *testing.T) {
		gpuRequest := resource.MustParse("1")
		requirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				resourceGPU: gpuRequest,
			},
		}

		accelerator := &core.GPUAccelerator{
			Device:      "amd-mi250",
			DeviceClass: core.GPUAccelerator_AMD_GPU,
		}

		expectedRequirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceName("amd.com/gpu"): gpuRequest,
			},
		}

		SanitizeGPUResourceRequirements(&requirements, accelerator)
		assert.EqualValues(t, expectedRequirements, requirements)
	})

	t.Run("HABANA_GAUDI device class", func(t *testing.T) {
		gpuRequest := resource.MustParse("1")
		requirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				resourceGPU: gpuRequest,
			},
		}

		accelerator := &core.GPUAccelerator{
			Device:      "habana-gaudi-dl1",
			DeviceClass: core.GPUAccelerator_HABANA_GAUDI,
		}

		expectedRequirements := v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceName("habana.ai/gaudi"): gpuRequest,
			},
		}

		SanitizeGPUResourceRequirements(&requirements, accelerator)
		assert.EqualValues(t, expectedRequirements, requirements)
	})
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
	inputReader.OnGetInputPath().Return(storage.DataReference("test-data-reference"))
	inputReader.OnGetInputPrefixPath().Return(storage.DataReference("test-data-reference-prefix"))
	inputReader.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)

	outputWriter := &mocks2.OutputWriter{}
	outputWriter.OnGetOutputPrefixPath().Return("")
	outputWriter.OnGetRawOutputPrefix().Return("")
	outputWriter.OnGetCheckpointPrefix().Return("/checkpoint")
	outputWriter.OnGetPreviousCheckpointsPrefix().Return("/prev")

	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskOverrides := mocks.TaskOverrides{}
	mockTaskOverrides.OnGetResources().Return(&v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: resource.MustParse("1024Mi"),
		},
	})
	mockTaskOverrides.OnGetExtendedResources().Return(nil)
	mockTaskExecMetadata.OnGetOverrides().Return(&mockTaskOverrides)
	mockTaskExecutionID := mocks.TaskExecutionID{}
	mockTaskExecutionID.OnGetID().Return(core.TaskExecutionIdentifier{})
	mockTaskExecutionID.OnGetGeneratedName().Return("gen_name")
	mockTaskExecMetadata.OnGetTaskExecutionID().Return(&mockTaskExecutionID)
	mockTaskExecMetadata.OnGetPlatformResources().Return(&v1.ResourceRequirements{})
	mockTaskExecMetadata.OnGetEnvironmentVariables().Return(map[string]string{
		"foo": "bar",
	})
	mockTaskExecMetadata.OnGetNamespace().Return("my-namespace")
	mockTaskExecMetadata.OnGetConsoleURL().Return("")

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(&mockTaskExecMetadata)
	tCtx.OnInputReader().Return(inputReader)
	tCtx.OnTaskReader().Return(taskReader)
	tCtx.OnOutputWriter().Return(outputWriter)

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
	mockTaskExecutionID.OnGetUniqueNodeID().Return("unique_node_id")
	mockTaskExecutionID.OnGetGeneratedName().Return("gen_name")
	mockTaskExecutionID.OnGetID().Return(core.TaskExecutionIdentifier{
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
	mockTaskExecMetadata.OnGetTaskExecutionID().Return(&mockTaskExecutionID)

	mockOverrides := mocks.TaskOverrides{}
	mockOverrides.OnGetResources().Return(resourceRequirements)
	mockTaskExecMetadata.OnGetOverrides().Return(&mockOverrides)
	mockTaskExecMetadata.OnGetPlatformResources().Return(platformResources)
	mockTaskExecMetadata.OnGetEnvironmentVariables().Return(nil)
	mockTaskExecMetadata.OnGetNamespace().Return("my-namespace")
	mockTaskExecMetadata.OnGetConsoleURL().Return(consoleURL)

	mockInputReader := mocks2.InputReader{}
	mockInputPath := storage.DataReference("s3://input/path")
	mockInputReader.OnGetInputPath().Return(mockInputPath)
	mockInputReader.OnGetInputPrefixPath().Return(mockInputPath)
	mockInputReader.On("Get", mock.Anything).Return(nil, nil)

	mockOutputPath := mocks2.OutputFilePaths{}
	mockOutputPathPrefix := storage.DataReference("s3://output/path")
	mockOutputPath.OnGetRawOutputPrefix().Return(mockOutputPathPrefix)
	mockOutputPath.OnGetOutputPrefixPath().Return(mockOutputPathPrefix)
	mockOutputPath.OnGetCheckpointPrefix().Return("/checkpoint")
	mockOutputPath.OnGetPreviousCheckpointsPrefix().Return("/prev")

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
	err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeAssignResources, container, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, container.Args, []string{"s3://output/path"})
	assert.EqualValues(t, container.Command, []string{"s3://input/path"})
	assert.Len(t, container.Resources.Limits, 3)
	assert.Len(t, container.Resources.Requests, 3)
	assert.Len(t, container.Env, 13)
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
		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, nil)
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
		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, nil)
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
		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, nil)
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

		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, nil)
		assert.NoError(t, err)
		assert.Equal(t, container.Resources.Requests[ResourceNvidiaGPU], overrideRequests[ResourceNvidiaGPU])
		assert.Equal(t, container.Resources.Limits[ResourceNvidiaGPU], overrideLimits[ResourceNvidiaGPU])
	})
	t.Run("ensure ExtendedResources.gpu_accelerator.device_class is respected when setting gpu resources", func(t *testing.T) {
		container := &v1.Container{
			Command: []string{
				"{{ .Input }}",
			},
			Args: []string{
				"{{ .OutputPrefix }}",
			},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					resourceGPU: resource.MustParse("2"),
				},
				Limits: v1.ResourceList{
					resourceGPU: resource.MustParse("2"),
				},
			},
		}

		tpuExtendedResources := &core.ExtendedResources{
			GpuAccelerator: &core.GPUAccelerator{
				Device:      "tpu-v4",
				DeviceClass: core.GPUAccelerator_GOOGLE_TPU,
			},
		}

		templateParameters := getTemplateParametersForTest(&v1.ResourceRequirements{}, &v1.ResourceRequirements{}, false, "")

		err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeMergeExistingResources, container, tpuExtendedResources)
		assert.NoError(t, err)

		// Verify generic "gpu" key is removed
		_, hasGenericGPU := container.Resources.Requests[resourceGPU]
		assert.False(t, hasGenericGPU)

		// Verify TPU resource is set correctly
		expectedTPU := resource.MustParse("2")
		assert.Equal(t, expectedTPU, container.Resources.Requests[v1.ResourceName("google.com/tpu")])
		assert.Equal(t, expectedTPU, container.Resources.Limits[v1.ResourceName("google.com/tpu")])
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
	err := AddFlyteCustomizationsToContainer(context.TODO(), templateParameters, ResourceCustomizationModeEnsureExistingResourcesInRange, container, nil)
	assert.NoError(t, err)

	assert.True(t, container.Resources.Requests.Cpu().Equal(resource.MustParse("10")))
	assert.True(t, container.Resources.Limits.Cpu().Equal(resource.MustParse("10")))
}

func TestAddFlyteCustomizationsToContainer_GPUResourceOverride(t *testing.T) {
	type testCase struct {
		name              string
		initialResources  v1.ResourceRequirements
		overrideResources v1.ResourceRequirements
		extendedResources *core.ExtendedResources
		customizationMode ResourceCustomizationMode
		expectedRequests  v1.ResourceList
		expectedLimits    v1.ResourceList
	}

	tests := []testCase{
		{
			name:             "override gpu: 1 translates to nvidia.com/gpu",
			initialResources: v1.ResourceRequirements{},
			overrideResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{resourceGPU: resource.MustParse("1")},
				Limits:   v1.ResourceList{resourceGPU: resource.MustParse("1")},
			},
			extendedResources: nil,
			customizationMode: ResourceCustomizationModeAssignResources,
			expectedRequests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
				ResourceNvidiaGPU: resource.MustParse("1"),
			},
			expectedLimits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
				ResourceNvidiaGPU: resource.MustParse("1"),
			},
		},
		{
			name:             "override gpu: 1 with extended resources for TPU",
			initialResources: v1.ResourceRequirements{},
			overrideResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{resourceGPU: resource.MustParse("1")},
			},
			extendedResources: &core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device:      "tpu-v4",
					DeviceClass: core.GPUAccelerator_GOOGLE_TPU,
				},
			},
			customizationMode: ResourceCustomizationModeAssignResources,
			expectedRequests: v1.ResourceList{
				v1.ResourceCPU:                    resource.MustParse("1"),
				v1.ResourceMemory:                 resource.MustParse("1Gi"),
				v1.ResourceName("google.com/tpu"): resource.MustParse("1"),
			},
			expectedLimits: v1.ResourceList{
				v1.ResourceCPU:                    resource.MustParse("1"),
				v1.ResourceMemory:                 resource.MustParse("1Gi"),
				v1.ResourceName("google.com/tpu"): resource.MustParse("1"),
			},
		},
		{
			name: "merge mode - override gpu on container with existing cpu/memory resources",
			initialResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("4Gi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("4"),
					v1.ResourceMemory: resource.MustParse("8Gi"),
				},
			},
			overrideResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{resourceGPU: resource.MustParse("2")},
				Limits:   v1.ResourceList{resourceGPU: resource.MustParse("2")},
			},
			extendedResources: nil,
			customizationMode: ResourceCustomizationModeMergeExistingResources,
			expectedRequests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("2"),
				v1.ResourceMemory: resource.MustParse("4Gi"),
				ResourceNvidiaGPU: resource.MustParse("2"),
			},
			expectedLimits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
				ResourceNvidiaGPU: resource.MustParse("2"),
			},
		},
		{
			name: "merge mode - override gpu replaces existing gpu in container",
			initialResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("4Gi"),
					resourceGPU:       resource.MustParse("1"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("4"),
					v1.ResourceMemory: resource.MustParse("8Gi"),
					resourceGPU:       resource.MustParse("1"),
				},
			},
			overrideResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{resourceGPU: resource.MustParse("4")},
				Limits:   v1.ResourceList{resourceGPU: resource.MustParse("4")},
			},
			extendedResources: nil,
			customizationMode: ResourceCustomizationModeMergeExistingResources,
			expectedRequests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("2"),
				v1.ResourceMemory: resource.MustParse("4Gi"),
				ResourceNvidiaGPU: resource.MustParse("4"),
			},
			expectedLimits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
				ResourceNvidiaGPU: resource.MustParse("4"),
			},
		},
		{
			name: "merge mode - override gpu with TPU on container with existing resources",
			initialResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("8"),
					v1.ResourceMemory: resource.MustParse("16Gi"),
					resourceGPU:       resource.MustParse("1"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("16"),
					v1.ResourceMemory: resource.MustParse("32Gi"),
					resourceGPU:       resource.MustParse("1"),
				},
			},
			overrideResources: v1.ResourceRequirements{
				Requests: v1.ResourceList{resourceGPU: resource.MustParse("8")},
				Limits:   v1.ResourceList{resourceGPU: resource.MustParse("8")},
			},
			extendedResources: &core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device:      "tpu-v5e",
					DeviceClass: core.GPUAccelerator_GOOGLE_TPU,
				},
			},
			customizationMode: ResourceCustomizationModeMergeExistingResources,
			expectedRequests: v1.ResourceList{
				v1.ResourceCPU:                    resource.MustParse("8"),
				v1.ResourceMemory:                 resource.MustParse("16Gi"),
				v1.ResourceName("google.com/tpu"): resource.MustParse("8"),
			},
			expectedLimits: v1.ResourceList{
				v1.ResourceCPU:                    resource.MustParse("16"),
				v1.ResourceMemory:                 resource.MustParse("32Gi"),
				v1.ResourceName("google.com/tpu"): resource.MustParse("8"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			container := &v1.Container{
				Command:   []string{"{{ .Input }}"},
				Args:      []string{"{{ .OutputPrefix }}"},
				Resources: tc.initialResources,
			}

			overrideResources := tc.overrideResources
			templateParameters := getTemplateParametersForTest(
				&overrideResources,
				&v1.ResourceRequirements{},
				false,
				"",
			)

			err := AddFlyteCustomizationsToContainer(
				context.TODO(),
				templateParameters,
				tc.customizationMode,
				container,
				tc.extendedResources,
			)
			assert.NoError(t, err)

			// Verify requests match exactly
			assert.Equal(t, len(tc.expectedRequests), len(container.Resources.Requests),
				"requests should have exactly %d resources", len(tc.expectedRequests))
			for resourceName, expectedQuantity := range tc.expectedRequests {
				actualQuantity := container.Resources.Requests[resourceName]
				assert.True(t, expectedQuantity.Equal(actualQuantity),
					"expected %s=%s in requests, got %s", resourceName, expectedQuantity.String(), actualQuantity.String())
			}

			// Verify limits match exactly
			assert.Equal(t, len(tc.expectedLimits), len(container.Resources.Limits),
				"limits should have exactly %d resources", len(tc.expectedLimits))
			for resourceName, expectedQuantity := range tc.expectedLimits {
				actualQuantity := container.Resources.Limits[resourceName]
				assert.True(t, expectedQuantity.Equal(actualQuantity),
					"expected %s=%s in limits, got %s", resourceName, expectedQuantity.String(), actualQuantity.String())
			}
		})
	}
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

	err := AddFlyteCustomizationsToContainer(context.TODO(), getTemplateParametersForTest(nil, nil, false, ""), ResourceCustomizationModeEnsureExistingResourcesInRange, container, nil)
	assert.NoError(t, err)

	assert.Len(t, container.EnvFrom, 2)
	assert.Equal(t, container.EnvFrom[0], configMapSource)
	assert.Equal(t, container.EnvFrom[1], secretSource)
}
