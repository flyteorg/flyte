package flytek8s

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/assert"
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
			{Name: core.Resources_STORAGE, Value: "1024Mi"},
			{Name: core.Resources_EPHEMERAL_STORAGE, Value: "1024Mi"},
		})

		assert.NoError(t, err)
		assert.NotEmpty(t, r)
		assert.NotNil(t, r[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("250m"), r[v1.ResourceCPU])
		assert.Equal(t, resource.MustParse("1"), r[ResourceNvidiaGPU])
		assert.Equal(t, resource.MustParse("1024Mi"), r[v1.ResourceMemory])
		assert.Equal(t, resource.MustParse("1024Mi"), r[v1.ResourceStorage])
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
	mockTaskExecMetadata.OnGetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: "service-account"},
	})
	result := GetServiceAccountNameFromTaskExecutionMetadata(&mockTaskExecMetadata)
	assert.Equal(t, "service-account", result)
}

func TestGetServiceAccountNameFromServiceAccount(t *testing.T) {
	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskExecMetadata.OnGetSecurityContext().Return(core.SecurityContext{})
	mockTaskExecMetadata.OnGetK8sServiceAccount().Return("service-account")
	result := GetServiceAccountNameFromTaskExecutionMetadata(&mockTaskExecMetadata)
	assert.Equal(t, "service-account", result)
}
