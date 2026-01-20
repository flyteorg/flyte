package flytek8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestToK8sEnvVar(t *testing.T) {
	e := ToK8sEnvVar([]*core.KeyValuePair{
		{Key: "k1", Value: "v1"},
		{Key: "k2", Value: "v2"},
	})

	assert.NotEmpty(t, e)
	assert.Equal(t, []v1.EnvVar{
		{Name: "k1", Value: "v1"},
		{Name: "k2", Value: "v2"},
	}, e)

	e = ToK8sEnvVar(nil)
	assert.Empty(t, e)
}

func TestToK8sResourceList(t *testing.T) {
	{
		r, err := ToK8sResourceList([]*core.Resources_ResourceEntry{
			{Name: core.Resources_CPU, Value: "250m"},
			{Name: core.Resources_GPU, Value: "1"},
			{Name: core.Resources_MEMORY, Value: "1024Mi"},
			{Name: core.Resources_EPHEMERAL_STORAGE, Value: "1024Mi"},
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, r)
		assert.NotNil(t, r[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("250m"), r[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("1"), r[resourceGPU])
		assert.Equal(t, resource.MustParse("1024Mi"), r[v1.ResourceMemory])
		assert.Equal(t, resource.MustParse("1024Mi"), r[v1.ResourceEphemeralStorage])
	}
	{
		r, err := ToK8sResourceList([]*core.Resources_ResourceEntry{})
		assert.NoError(t, err)
		assert.Empty(t, r)
	}
	{
		_, err := ToK8sResourceList([]*core.Resources_ResourceEntry{
			{Name: core.Resources_CPU, Value: "250x"},
		})
		assert.Error(t, err)
	}

}

func TestToK8sResourceRequirements(t *testing.T) {

	{
		r, err := ToK8sResourceRequirements(nil)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Empty(t, r.Limits)
		assert.Empty(t, r.Requests)
	}
	{
		r, err := ToK8sResourceRequirements(&core.Resources{
			Requests: nil,
			Limits:   nil,
		})
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Empty(t, r.Limits)
		assert.Empty(t, r.Requests)
	}
	{
		r, err := ToK8sResourceRequirements(&core.Resources{
			Requests: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "250m"},
			},
			Limits: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "1024m"},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Equal(t, resource.MustParse("250m"), r.Requests[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("1024m"), r.Limits[v1.ResourceCPU])
	}
	{
		_, err := ToK8sResourceRequirements(&core.Resources{
			Requests: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "blah"},
			},
			Limits: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "1024m"},
			},
		})
		assert.Error(t, err)
	}
	{
		_, err := ToK8sResourceRequirements(&core.Resources{
			Requests: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "250m"},
			},
			Limits: []*core.Resources_ResourceEntry{
				{Name: core.Resources_CPU, Value: "blah"},
			},
		})
		assert.Error(t, err)
	}
}

func TestGetServiceAccountNameFromTaskExecutionMetadata(t *testing.T) {
	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskExecMetadata.EXPECT().GetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: "service-account"},
	})
	result := GetServiceAccountNameFromTaskExecutionMetadata(&mockTaskExecMetadata)
	assert.Equal(t, "service-account", result)
}

func TestGetServiceAccountNameFromServiceAccount(t *testing.T) {
	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskExecMetadata.EXPECT().GetSecurityContext().Return(core.SecurityContext{})
	mockTaskExecMetadata.EXPECT().GetK8sServiceAccount().Return("service-account")
	result := GetServiceAccountNameFromTaskExecutionMetadata(&mockTaskExecMetadata)
	assert.Equal(t, "service-account", result)
}

func TestGetNormalizedAcceleratorDevice(t *testing.T) {
	// Setup config with AcceleratorDevices mapping
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		AcceleratorDevices: map[string]string{
			"A100":   "nvidia-tesla-a100",
			"H100":   "nvidia-h100",
			"T4":     "nvidia-tesla-t4",
			"V5E":    "tpu-v5-lite-podslice",
			"V5P":    "tpu-v5p-slice",
			"INF2":   "aws-neuron-inf2",
			"TRN1":   "aws-neuron-trn1",
			"MI300X": "amd-mi300x",
			"MI250X": "amd-mi250x",
			"DL1":    "habana-gaudi-dl1",
		},
	}))

	t.Run("NVIDIA GPU normalization", func(t *testing.T) {
		assert.Equal(t, "nvidia-tesla-a100", GetNormalizedAcceleratorDevice("A100"))
		assert.Equal(t, "nvidia-h100", GetNormalizedAcceleratorDevice("H100"))
		assert.Equal(t, "nvidia-tesla-t4", GetNormalizedAcceleratorDevice("T4"))
	})

	t.Run("Google TPU normalization", func(t *testing.T) {
		assert.Equal(t, "tpu-v5-lite-podslice", GetNormalizedAcceleratorDevice("V5E"))
		assert.Equal(t, "tpu-v5p-slice", GetNormalizedAcceleratorDevice("V5P"))
	})

	t.Run("AWS Neuron normalization", func(t *testing.T) {
		assert.Equal(t, "aws-neuron-inf2", GetNormalizedAcceleratorDevice("INF2"))
		assert.Equal(t, "aws-neuron-trn1", GetNormalizedAcceleratorDevice("TRN1"))
	})

	t.Run("AMD GPU normalization", func(t *testing.T) {
		assert.Equal(t, "amd-mi300x", GetNormalizedAcceleratorDevice("MI300X"))
		assert.Equal(t, "amd-mi250x", GetNormalizedAcceleratorDevice("MI250X"))
	})

	t.Run("Habana Gaudi normalization", func(t *testing.T) {
		assert.Equal(t, "habana-gaudi-dl1", GetNormalizedAcceleratorDevice("DL1"))
	})

	t.Run("case insensitivity", func(t *testing.T) {
		assert.Equal(t, "nvidia-tesla-a100", GetNormalizedAcceleratorDevice("a100"))
		assert.Equal(t, "nvidia-tesla-a100", GetNormalizedAcceleratorDevice("A100"))
		assert.Equal(t, "nvidia-h100", GetNormalizedAcceleratorDevice("h100"))
		assert.Equal(t, "tpu-v5-lite-podslice", GetNormalizedAcceleratorDevice("v5e"))
		assert.Equal(t, "aws-neuron-inf2", GetNormalizedAcceleratorDevice("inf2"))
	})

	t.Run("unmapped device fallback", func(t *testing.T) {
		assert.Equal(t, "custom-device", GetNormalizedAcceleratorDevice("custom-device"))
		assert.Equal(t, "unknown-gpu", GetNormalizedAcceleratorDevice("unknown-gpu"))
	})

	t.Run("empty device", func(t *testing.T) {
		assert.Equal(t, "", GetNormalizedAcceleratorDevice(""))
	})
}

func TestGetAcceleratorResourceName(t *testing.T) {
	t.Run("returns device class specific resource name", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName: "nvidia.com/gpu",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{
				"NVIDIA_GPU": {
					ResourceName: "nvidia.com/gpu",
				},
				"GOOGLE_TPU": {
					ResourceName: "google.com/tpu",
				},
				"AMAZON_NEURON": {
					ResourceName: "aws.amazon.com/neuron",
				},
				"AMD_GPU": {
					ResourceName: "amd.com/gpu",
				},
			},
		}))

		// Test NVIDIA GPU
		result := getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_NVIDIA_GPU,
		})
		assert.Equal(t, v1.ResourceName("nvidia.com/gpu"), result)

		// Test Google TPU
		result = getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_GOOGLE_TPU,
		})
		assert.Equal(t, v1.ResourceName("google.com/tpu"), result)

		// Test Amazon Neuron
		result = getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_AMAZON_NEURON,
		})
		assert.Equal(t, v1.ResourceName("aws.amazon.com/neuron"), result)

		// Test AMD GPU
		result = getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_AMD_GPU,
		})
		assert.Equal(t, v1.ResourceName("amd.com/gpu"), result)
	})

	t.Run("falls back to legacy GpuResourceName when device class not configured", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName:          "nvidia.com/gpu",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{},
		}))

		result := getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_NVIDIA_GPU,
		})
		assert.Equal(t, v1.ResourceName("nvidia.com/gpu"), result)
	})

	t.Run("falls back to legacy GpuResourceName when device class config not found", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName: "custom.gpu.resource",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{
				"GOOGLE_TPU": {
					ResourceName: "google.com/tpu",
				},
			},
		}))

		// Use NVIDIA_GPU (default, value 0) but it's not in the config, so should fallback
		result := getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_NVIDIA_GPU,
		})
		assert.Equal(t, v1.ResourceName("custom.gpu.resource"), result)
	})

	t.Run("falls back to legacy GpuResourceName when accelerator is nil", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName:          "nvidia.com/gpu",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{},
		}))

		result := getAcceleratorResourceName(nil)
		assert.Equal(t, v1.ResourceName("nvidia.com/gpu"), result)
	})

	t.Run("uses device class config even when resource name is empty string", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName: "nvidia.com/gpu",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{
				"NVIDIA_GPU": {
					ResourceName: "", // Empty - should fallback
				},
			},
		}))

		result := getAcceleratorResourceName(&core.GPUAccelerator{
			DeviceClass: core.GPUAccelerator_NVIDIA_GPU,
		})
		// Should fallback to global GpuResourceName when device class resource name is empty
		assert.Equal(t, v1.ResourceName("nvidia.com/gpu"), result)
	})
}

func TestGetAllAcceleratorResourceNames(t *testing.T) {
	t.Run("returns all configured accelerator resource names", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName: "nvidia.com/gpu",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{
				"NVIDIA_GPU": {
					ResourceName: "nvidia.com/gpu",
				},
				"GOOGLE_TPU": {
					ResourceName: "google.com/tpu",
				},
				"AMAZON_NEURON": {
					ResourceName: "aws.amazon.com/neuron",
				},
				"AMD_GPU": {
					ResourceName: "amd.com/gpu",
				},
				"HABANA_GAUDI": {
					ResourceName: "habana.ai/gaudi",
				},
			},
		}))

		result := getAllAcceleratorResourceNames()

		// Should include legacy GpuResourceName
		assert.Contains(t, result, v1.ResourceName("nvidia.com/gpu"))
		// Should include all configured accelerator resource names
		assert.Contains(t, result, v1.ResourceName("google.com/tpu"))
		assert.Contains(t, result, v1.ResourceName("aws.amazon.com/neuron"))
		assert.Contains(t, result, v1.ResourceName("amd.com/gpu"))
		assert.Contains(t, result, v1.ResourceName("habana.ai/gaudi"))

		// Should have exactly 5 unique resource names (nvidia.com/gpu appears in both legacy and new map)
		assert.Equal(t, 5, len(result))
	})

	t.Run("includes legacy GpuResourceName for backward compatibility", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName:          "custom.gpu.resource",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{},
		}))

		result := getAllAcceleratorResourceNames()

		assert.Contains(t, result, v1.ResourceName("custom.gpu.resource"))
		assert.Equal(t, 1, len(result))
	})

	t.Run("ensures uniqueness when legacy and new map overlap", func(t *testing.T) {
		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			GpuResourceName: "nvidia.com/gpu",
			AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{
				"NVIDIA_GPU": {
					ResourceName: "nvidia.com/gpu",
				},
				"GOOGLE_TPU": {
					ResourceName: "google.com/tpu",
				},
			},
		}))

		result := getAllAcceleratorResourceNames()

		// Should deduplicate nvidia.com/gpu
		assert.Contains(t, result, v1.ResourceName("nvidia.com/gpu"))
		assert.Contains(t, result, v1.ResourceName("google.com/tpu"))
		assert.Equal(t, 2, len(result))
	})
}

func TestPodRequiresAccelerator(t *testing.T) {
	// Setup config with multiple accelerator types
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		GpuResourceName: "nvidia.com/gpu",
		AcceleratorDeviceClasses: map[string]config.AcceleratorDeviceClassConfig{
			"NVIDIA_GPU": {
				ResourceName: "nvidia.com/gpu",
			},
			"GOOGLE_TPU": {
				ResourceName: "google.com/tpu",
			},
			"AMAZON_NEURON": {
				ResourceName: "aws.amazon.com/neuron",
			},
			"AMD_GPU": {
				ResourceName: "amd.com/gpu",
			},
			"HABANA_GAUDI": {
				ResourceName: "habana.ai/gaudi",
			},
		},
	}))

	t.Run("pod with NVIDIA GPU resources returns true", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"nvidia.com/gpu": resource.MustParse("1"),
						},
					},
				},
			},
		}
		assert.True(t, podRequiresAccelerator(podSpec))
	})

	t.Run("pod with Google TPU resources returns true", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"google.com/tpu": resource.MustParse("4"),
						},
					},
				},
			},
		}
		assert.True(t, podRequiresAccelerator(podSpec))
	})

	t.Run("pod with AWS Neuron resources returns true", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"aws.amazon.com/neuron": resource.MustParse("2"),
						},
					},
				},
			},
		}
		assert.True(t, podRequiresAccelerator(podSpec))
	})

	t.Run("pod with AMD GPU resources returns true", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"amd.com/gpu": resource.MustParse("1"),
						},
					},
				},
			},
		}
		assert.True(t, podRequiresAccelerator(podSpec))
	})

	t.Run("pod with Habana Gaudi resources returns true", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"habana.ai/gaudi": resource.MustParse("1"),
						},
					},
				},
			},
		}
		assert.True(t, podRequiresAccelerator(podSpec))
	})

	t.Run("pod with no accelerator resources returns false", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("1"),
							v1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		}
		assert.False(t, podRequiresAccelerator(podSpec))
	})

	t.Run("pod with only CPU and memory returns false", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("1"),
							v1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		}
		assert.False(t, podRequiresAccelerator(podSpec))
	})

	t.Run("multiple containers with one having accelerator returns true", func(t *testing.T) {
		podSpec := &v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
				{
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"nvidia.com/gpu": resource.MustParse("1"),
						},
					},
				},
			},
		}
		assert.True(t, podRequiresAccelerator(podSpec))
	})

	t.Run("empty pod spec returns false", func(t *testing.T) {
		podSpec := &v1.PodSpec{}
		assert.False(t, podRequiresAccelerator(podSpec))
	})
}
