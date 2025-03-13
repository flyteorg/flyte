package flytek8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
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
		}, nil)

		assert.NoError(t, err)
		assert.NotEmpty(t, r)
		assert.NotNil(t, r[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("250m"), r[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("1"), r[ResourceNvidiaGPU])
		assert.Equal(t, resource.MustParse("1024Mi"), r[v1.ResourceMemory])
		assert.Equal(t, resource.MustParse("1024Mi"), r[v1.ResourceEphemeralStorage])
	}
	{
		r, err := ToK8sResourceList([]*core.Resources_ResourceEntry{}, nil)
		assert.NoError(t, err)
		assert.Empty(t, r)
	}
	{
		_, err := ToK8sResourceList([]*core.Resources_ResourceEntry{
			{Name: core.Resources_CPU, Value: "250x"},
		}, nil)
		assert.Error(t, err)
	}
	{
		// Test with non-nil onOOMConfig
		mockOnOOMConfig := &mocks.OnOOMConfig{}
		mockOnOOMConfig.EXPECT().GetExponent().Return(uint32(2))
		mockOnOOMConfig.EXPECT().GetFactor().Return(float32(2.0))
		mockOnOOMConfig.EXPECT().GetLimit().Return("4096Mi")

		r, err := ToK8sResourceList([]*core.Resources_ResourceEntry{
			{Name: core.Resources_MEMORY, Value: "1024Mi"},
		}, mockOnOOMConfig)

		assert.NoError(t, err)
		assert.NotEmpty(t, r)
		// Memory should be multiplied by 2^2 = 4
		// Compare the raw values by converting to bytes
		expectedQty := resource.MustParse("4096Mi")
		actualQty := r[v1.ResourceMemory] // Extract value from map

		assert.Equal(t, 0, actualQty.Cmp(expectedQty),
			"Resource values do not match: expected %s, got %s", expectedQty.String(), actualQty.String())
		// Check if resource.Quantity's string matches its numerical value
		assert.Equal(t, actualQty.String(), "4294967296")
	}
	{
		mockOnOOMConfig := &mocks.OnOOMConfig{}
		mockOnOOMConfig.EXPECT().GetExponent().Return(uint32(2))
		mockOnOOMConfig.EXPECT().GetFactor().Return(float32(2.0))
		mockOnOOMConfig.EXPECT().GetLimit().Return("4096Mi")

		// Test with a value that would exceed the limit
		r, err := ToK8sResourceList([]*core.Resources_ResourceEntry{
			{Name: core.Resources_MEMORY, Value: "2048Mi"},
		}, mockOnOOMConfig)

		assert.NoError(t, err)
		assert.NotEmpty(t, r)
		// 2048Mi * 4 = 8192Mi = 8Gi, but limit is 4096Mi
		expectedQty := resource.MustParse("4096Mi")
		actualQty := r[v1.ResourceMemory] // Extract value from map

		assert.Equal(t, actualQty, expectedQty,
			"Resource values do not match: expected %s, got %s", expectedQty.String(), actualQty.String())
	}
}

func TestToK8sResourceRequirements(t *testing.T) {

	{
		r, err := ToK8sResourceRequirements(nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.Empty(t, r.Limits)
		assert.Empty(t, r.Requests)
	}
	{
		r, err := ToK8sResourceRequirements(&core.Resources{
			Requests: nil,
			Limits:   nil,
		}, nil)
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
		}, nil)
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
		}, nil)
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
		}, nil)
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
