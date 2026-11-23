package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
)

func TestTaskActionForCrossNamespacePod(t *testing.T) {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "gpu-workers", Annotations: map[string]string{
		flytek8s.TaskActionNamespaceAnnotation: "flyte", flytek8s.TaskActionNameAnnotation: "task-a0",
	}}}
	requests := taskActionForCrossNamespacePod(context.Background(), pod)
	require.Len(t, requests, 1)
	require.Equal(t, client.ObjectKey{Namespace: "flyte", Name: "task-a0"}, requests[0].NamespacedName)
	delete(pod.Annotations, flytek8s.TaskActionNameAnnotation)
	require.Empty(t, taskActionForCrossNamespacePod(context.Background(), pod))
	pod.Annotations = nil
	require.Empty(t, taskActionForCrossNamespacePod(context.Background(), pod))
}
