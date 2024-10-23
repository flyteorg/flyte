package v1alpha1

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestKind(t *testing.T) {
	kind := "test kind"
	got := Kind(kind)
	want := SchemeGroupVersion.WithKind(kind).GroupKind()
	assert.Equal(t, got, want)
}

func TestResource(t *testing.T) {
	resource := "test resource"
	got := Resource(resource)
	want := SchemeGroupVersion.WithResource(resource).GroupResource()
	assert.Equal(t, got, want)
}

func Test_addKnownTypes(t *testing.T) {
	scheme := runtime.NewScheme()
	err := addKnownTypes(scheme)
	assert.Nil(t, err)
}
