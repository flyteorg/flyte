package k8s

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFinalize(t *testing.T) {
	ctx := context.Background()

	tCtx := getMockTaskExecutionContext(ctx, 0)
	kubeClient := mocks.KubeClient{}
	kubeClient.OnGetClient().Return(mocks.NewFakeKubeClient())

	resourceManager := mocks.ResourceManager{}
	podTemplate, _, _ := FlyteArrayJobToK8sPodTemplate(ctx, tCtx, "")
	pod := addPodFinalizer(&podTemplate)
	pod.Name = formatSubTaskName(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), 1, 1)
	assert.Equal(t, "notfound-1-1", pod.Name)
	assert.NoError(t, kubeClient.GetClient().Create(ctx, pod))

	resourceManager.OnReleaseResourceMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)
	tCtx.OnResourceManager().Return(&resourceManager)

	config := Config{
		MaxArrayJobSize: 100,
		ResourceConfig: ResourceConfig{
			PrimaryLabel: "p",
			Limit:        10,
		},
	}

	retryAttemptsArray, err := bitarray.NewCompactArray(2, 1)
	assert.NoError(t, err)

	state := core.State{
		RetryAttempts: retryAttemptsArray,
	}

	task := &Task{
		State:    &state,
		Config:   &config,
		ChildIdx: 1,
	}

	err = task.Finalize(ctx, tCtx, &kubeClient)
	assert.NoError(t, err)
}

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
