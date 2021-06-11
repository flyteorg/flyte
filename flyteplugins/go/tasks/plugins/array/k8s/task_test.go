package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetTaskContainerIndex(t *testing.T) {
	t.Run("test container target", func(t *testing.T) {
		pod := &v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container",
					},
				},
			},
		}
		index, err := getTaskContainerIndex(pod)
		assert.NoError(t, err)
		assert.Equal(t, 0, index)
	})
	t.Run("test missing primary container annotation", func(t *testing.T) {
		pod := &v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container",
					},
					{
						Name: "container b",
					},
				},
			},
		}
		_, err := getTaskContainerIndex(pod)
		assert.EqualError(t, err, "[POD_TEMPLATE_FAILED] Expected a specified primary container key when building an array job with a K8sPod spec target")
	})
	t.Run("test get primary container index", func(t *testing.T) {
		pod := &v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container a",
					},
					{
						Name: "container b",
					},
					{
						Name: "container c",
					},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					primaryContainerKey: "container c",
				},
			},
		}
		index, err := getTaskContainerIndex(pod)
		assert.NoError(t, err)
		assert.Equal(t, 2, index)
	})
	t.Run("specified primary container doesn't exist", func(t *testing.T) {
		pod := &v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name: "container a",
					},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					primaryContainerKey: "container c",
				},
			},
		}
		_, err := getTaskContainerIndex(pod)
		assert.EqualError(t, err, "[POD_TEMPLATE_FAILED] Couldn't find any container matching the primary container key when building an array job with a K8sPod spec target")
	})
}
