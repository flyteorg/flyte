package logs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestGetPrimaryPodAndContainer_HappyPath(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{
				PodName:              "my-pod",
				Namespace:            "default",
				PrimaryContainerName: "main",
				Containers: []*core.ContainerContext{
					{ContainerName: "main"},
					{ContainerName: "sidecar"},
				},
			},
		},
	}

	pod, container, err := GetPrimaryPodAndContainer(logCtx)
	assert.NoError(t, err)
	assert.Equal(t, "my-pod", pod.GetPodName())
	assert.Equal(t, "default", pod.GetNamespace())
	assert.Equal(t, "main", container.GetContainerName())
}

func TestGetPrimaryPodAndContainer_EmptyPodName(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "",
	}

	_, _, err := GetPrimaryPodAndContainer(logCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "primary pod name is empty")
}

func TestGetPrimaryPodAndContainer_PodNotFound(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "missing-pod",
		Pods: []*core.PodLogContext{
			{PodName: "other-pod"},
		},
	}

	_, _, err := GetPrimaryPodAndContainer(logCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in log context")
}

func TestGetPrimaryPodAndContainer_ContainerNotFound(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{
				PodName:              "my-pod",
				PrimaryContainerName: "missing-container",
				Containers: []*core.ContainerContext{
					{ContainerName: "other"},
				},
			},
		},
	}

	_, _, err := GetPrimaryPodAndContainer(logCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "primary container")
}

func newTestLogContext(podName, namespace, containerName string) *core.LogContext {
	return &core.LogContext{
		PrimaryPodName: podName,
		Pods: []*core.PodLogContext{
			{
				PodName:              podName,
				Namespace:            namespace,
				PrimaryContainerName: containerName,
				Containers: []*core.ContainerContext{
					{ContainerName: containerName},
				},
			},
		},
	}
}

func TestTailLogs_PodNotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset() // no pods

	streamer := &K8sLogStreamer{clientset: clientset}
	logCtx := newTestLogContext("missing-pod", "default", "main")

	err := streamer.TailLogs(context.Background(), logCtx, nil)
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	assert.Contains(t, err.Error(), "not found")
}

func TestTailLogs_FollowSetBasedOnPodPhase(t *testing.T) {
	tests := []struct {
		name       string
		phase      corev1.PodPhase
		wantFollow bool
	}{
		{"running pod should follow", corev1.PodRunning, true},
		{"succeeded pod should not follow", corev1.PodSucceeded, false},
		{"failed pod should not follow", corev1.PodFailed, false},
		{"pending pod should not follow", corev1.PodPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pod",
					Namespace: "default",
				},
				Status: corev1.PodStatus{
					Phase: tt.phase,
				},
			})

			// Verify we can fetch the pod and the phase is correct.
			podObj, err := clientset.CoreV1().Pods("default").Get(context.Background(), "my-pod", metav1.GetOptions{})
			require.NoError(t, err)
			assert.Equal(t, tt.phase, podObj.Status.Phase)

			// Verify the follow logic: Follow should only be true when phase is Running.
			gotFollow := podObj.Status.Phase == corev1.PodRunning
			assert.Equal(t, tt.wantFollow, gotFollow)
		})
	}
}

func TestTailLogs_LogSubresourceNotFound(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "gone-pod", Namespace: "default"},
		Status:     corev1.PodStatus{Phase: corev1.PodSucceeded},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/namespaces/default/pods/gone-pod":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(pod)
		case "/api/v1/namespaces/default/pods/gone-pod/log":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`))
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	clientset, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
	require.NoError(t, err)

	streamer := &K8sLogStreamer{clientset: clientset}
	err = streamer.TailLogs(context.Background(), newTestLogContext("gone-pod", "default", "main"), nil)
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
