package flytek8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestApplyResourceOverrides_OverrideCpu(t *testing.T) {
	cpuRequest := resource.MustParse("1")
	overrides := ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: cpuRequest,
		},
	})
	assert.EqualValues(t, cpuRequest, overrides.Requests[v1.ResourceCPU])
	assert.EqualValues(t, cpuRequest, overrides.Limits[v1.ResourceCPU])

	cpuLimit := resource.MustParse("2")
	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU: cpuRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU: cpuLimit,
		},
	})
	assert.EqualValues(t, cpuRequest, overrides.Requests[v1.ResourceCPU])
	assert.EqualValues(t, cpuLimit, overrides.Limits[v1.ResourceCPU])

	// request equals limit if not set
	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU: cpuLimit,
		},
	})
	assert.EqualValues(t, cpuLimit, overrides.Requests[v1.ResourceCPU])
	assert.EqualValues(t, cpuLimit, overrides.Limits[v1.ResourceCPU])
}

func TestApplyResourceOverrides_OverrideMemory(t *testing.T) {
	memoryRequest := resource.MustParse("1")
	overrides := ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: memoryRequest,
		},
	})
	assert.EqualValues(t, memoryRequest, overrides.Requests[v1.ResourceMemory])
	assert.EqualValues(t, memoryRequest, overrides.Limits[v1.ResourceMemory])

	memoryLimit := resource.MustParse("2")
	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: memoryRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceMemory: memoryLimit,
		},
	})
	assert.EqualValues(t, memoryRequest, overrides.Requests[v1.ResourceMemory])
	assert.EqualValues(t, memoryLimit, overrides.Limits[v1.ResourceMemory])

	// request equals limit if not set
	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceMemory: memoryLimit,
		},
	})
	assert.EqualValues(t, memoryLimit, overrides.Requests[v1.ResourceMemory])
	assert.EqualValues(t, memoryLimit, overrides.Limits[v1.ResourceMemory])
}

func TestApplyResourceOverrides_OverrideEphemeralStorage(t *testing.T) {
	ephemeralStorageRequest := resource.MustParse("1")
	overrides := ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageRequest,
		},
	})
	assert.EqualValues(t, ephemeralStorageRequest, overrides.Requests[v1.ResourceEphemeralStorage])
	assert.EqualValues(t, ephemeralStorageRequest, overrides.Limits[v1.ResourceEphemeralStorage])

	ephemeralStorageLimit := resource.MustParse("2")
	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageRequest,
		},
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageLimit,
		},
	})
	assert.EqualValues(t, ephemeralStorageRequest, overrides.Requests[v1.ResourceEphemeralStorage])
	assert.EqualValues(t, ephemeralStorageLimit, overrides.Limits[v1.ResourceEphemeralStorage])

	// request equals limit if not set
	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceEphemeralStorage: ephemeralStorageLimit,
		},
	})
	assert.EqualValues(t, ephemeralStorageLimit, overrides.Requests[v1.ResourceEphemeralStorage])
}

func TestApplyResourceOverrides_RemoveStorage(t *testing.T) {
	requestedResourceQuantity := resource.MustParse("1")
	overrides := ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceStorage:          requestedResourceQuantity,
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceCPU:              requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
		Limits: v1.ResourceList{
			v1.ResourceStorage:          requestedResourceQuantity,
			v1.ResourceMemory:           requestedResourceQuantity,
			v1.ResourceEphemeralStorage: requestedResourceQuantity,
		},
	})
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
	overrides := ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Requests: v1.ResourceList{
			resourceGPU: gpuRequest,
		},
	})
	assert.EqualValues(t, gpuRequest, overrides.Requests[ResourceNvidiaGPU])

	overrides = ApplyResourceOverrides(context.Background(), v1.ResourceRequirements{
		Limits: v1.ResourceList{
			resourceGPU: gpuRequest,
		},
	})
	assert.EqualValues(t, gpuRequest, overrides.Limits[ResourceNvidiaGPU])
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
			v1.ResourceStorage:          requestedResourceQuantity,
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
			v1.ResourceStorage:          requestedResourceQuantity,
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
