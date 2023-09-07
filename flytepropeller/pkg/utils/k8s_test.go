package utils

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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
