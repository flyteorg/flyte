package utils

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
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

func TestGetProtoTime(t *testing.T) {
	assert.NotNil(t, GetProtoTime(nil))
	n := time.Now()
	nproto, err := ptypes.TimestampProto(n)
	assert.NoError(t, err)
	assert.Equal(t, nproto, GetProtoTime(&metav1.Time{Time: n}))
}

func TestGetWorkflowIDFromOwner(t *testing.T) {
	tests := []struct {
		name          string
		reference     *metav1.OwnerReference
		namespace     string
		expectedOwner string
		expectedErr   error
	}{
		{"nilReference", nil, "", "", NotTheOwnerError},
		{"badReference", &metav1.OwnerReference{Kind: "x"}, "", "", NotTheOwnerError},
		{"wfReference", &metav1.OwnerReference{Kind: v1alpha1.FlyteWorkflowKind, Name: "x"}, "ns", "ns/x", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o, e := GetWorkflowIDFromOwner(test.reference, test.namespace)
			assert.Equal(t, test.expectedOwner, o)
			assert.Equal(t, test.expectedErr, e)
		})
	}
}

func TestSanitizeLabelValue(t *testing.T) {
	assert.Equal(t, "a-b-c", SanitizeLabelValue("a.b.c"))
	assert.Equal(t, "a-b-c", SanitizeLabelValue("a.B.c"))
	assert.Equal(t, "a-9-c", SanitizeLabelValue("a.9.c"))
	assert.Equal(t, "a-b-c", SanitizeLabelValue("a-b-c"))
	assert.Equal(t, "a-b-c", SanitizeLabelValue("a-b-c/"))
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SanitizeLabelValue("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SanitizeLabelValue("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa."))
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SanitizeLabelValue("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab"))
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SanitizeLabelValue("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa."))
}
