package manager

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flytepropeller/manager/shardstrategy"
	"github.com/flyteorg/flytepropeller/manager/shardstrategy/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	podTemplate = &v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "0",
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo": "bar",
				},
				Labels: map[string]string{
					"app": "foo",
					"bar": "baz",
				},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{
						Command: []string{"flytepropeller"},
						Args:    []string{"--config", "/etc/flyte/config/*.yaml"},
					},
				},
			},
		},
	}
)

func createShardStrategy(podCount int) shardstrategy.ShardStrategy {
	shardStrategy := mocks.ShardStrategy{}
	shardStrategy.OnGetPodCount().Return(podCount)
	shardStrategy.OnHashCode().Return(0, nil)
	shardStrategy.OnUpdatePodSpecMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

	return &shardStrategy
}

func TestCreatePods(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		shardStrategy shardstrategy.ShardStrategy
	}{
		{"2", createShardStrategy(2)},
		{"3", createShardStrategy(3)},
		{"4", createShardStrategy(4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			scope := promutils.NewScope(fmt.Sprintf("create_%s", tt.name))
			kubeClient := fake.NewSimpleClientset(podTemplate)

			manager := Manager{
				kubeClient:     kubeClient,
				metrics:        newManagerMetrics(scope),
				podApplication: "flytepropeller",
				shardStrategy:  tt.shardStrategy,
			}

			// ensure no pods are "running"
			kubePodsClient := kubeClient.CoreV1().Pods("")
			pods, err := kubePodsClient.List(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, 0, len(pods.Items))

			// create all pods and validate state
			err = manager.createPods(ctx)
			assert.NoError(t, err)

			pods, err = kubePodsClient.List(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.shardStrategy.GetPodCount(), len(pods.Items))

			for _, pod := range pods.Items {
				assert.Equal(t, pod.Annotations["foo"], "bar")
				assert.Equal(t, pod.Labels["app"], "flytepropeller")
				assert.Equal(t, pod.Labels["bar"], "baz")
			}

			// execute again to ensure no new pods are created
			err = manager.createPods(ctx)
			assert.NoError(t, err)

			pods, err = kubePodsClient.List(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.shardStrategy.GetPodCount(), len(pods.Items))
		})
	}
}

func TestUpdatePods(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		shardStrategy shardstrategy.ShardStrategy
	}{
		{"2", createShardStrategy(2)},
		{"3", createShardStrategy(3)},
		{"4", createShardStrategy(4)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()
			scope := promutils.NewScope(fmt.Sprintf("update_%s", tt.name))

			initObjects := []runtime.Object{podTemplate}
			for i := 0; i < tt.shardStrategy.GetPodCount(); i++ {
				initObjects = append(initObjects, &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							podTemplateResourceVersion: "1",
							shardConfigHash:            "1",
						},
						Labels: map[string]string{
							"app": "flytepropeller",
						},
						Name: fmt.Sprintf("flytepropeller-%d", i),
					},
				})
			}

			kubeClient := fake.NewSimpleClientset(initObjects...)

			manager := Manager{
				kubeClient:     kubeClient,
				metrics:        newManagerMetrics(scope),
				podApplication: "flytepropeller",
				shardStrategy:  tt.shardStrategy,
			}

			// ensure all pods are "running"
			kubePodsClient := kubeClient.CoreV1().Pods("")
			pods, err := kubePodsClient.List(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.shardStrategy.GetPodCount(), len(pods.Items))
			for _, pod := range pods.Items {
				assert.Equal(t, "1", pod.ObjectMeta.Annotations[podTemplateResourceVersion])
			}

			// create all pods and validate state
			err = manager.createPods(ctx)
			assert.NoError(t, err)

			pods, err = kubePodsClient.List(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.shardStrategy.GetPodCount(), len(pods.Items))
			for _, pod := range pods.Items {
				assert.Equal(t, podTemplate.ObjectMeta.ResourceVersion, pod.ObjectMeta.Annotations[podTemplateResourceVersion])
			}
		})
	}
}

func TestGetPodNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		shardStrategy shardstrategy.ShardStrategy
		podCount      int
	}{
		{"2", createShardStrategy(2), 2},
		{"3", createShardStrategy(3), 3},
		{"4", createShardStrategy(4), 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := Manager{
				podApplication: "flytepropeller",
				shardStrategy:  tt.shardStrategy,
			}

			assert.Equal(t, tt.podCount, len(manager.getPodNames()))
		})
	}
}
