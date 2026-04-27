package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	appk8s "github.com/flyteorg/flyte/v2/app/internal/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/k8s/podlogs"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
)

// knativeQueueProxy is the Knative sidecar container injected into every
// KService pod. We always stream from the user container instead.
const knativeQueueProxy = "queue-proxy"

const defaultInitialLines = int64(1000)

// K8sAppLogStreamer streams logs from the pod backing an app replica.
type K8sAppLogStreamer struct {
	clientset kubernetes.Interface
}

// NewK8sAppLogStreamer creates a K8sAppLogStreamer from a Kubernetes REST config.
// It clears the timeout so that long-lived log streams are not interrupted.
func NewK8sAppLogStreamer(k8sConfig *rest.Config) (*K8sAppLogStreamer, error) {
	cfg := rest.CopyConfig(k8sConfig)
	cfg.Timeout = 0
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	return &K8sAppLogStreamer{clientset: clientset}, nil
}

// TailLogs streams log lines for a replica's pod.
func (s *K8sAppLogStreamer) TailLogs(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier, send func(*flyteapp.LogLines) error) error {
	ns := appk8s.AppNamespace
	podName := replicaID.GetName()

	pod, err := s.clientset.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("pod %s not found in namespace %s", podName, ns))
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pod: %w", err))
	}

	containerName := pickUserContainer(pod)
	if containerName == "" {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("no user container found in pod %s/%s", ns, podName))
	}

	tailLines := defaultInitialLines
	opts := &corev1.PodLogOptions{
		Container:  containerName,
		Follow:     pod.Status.Phase == corev1.PodRunning,
		Timestamps: true,
		TailLines:  &tailLines,
	}

	// Detach from the inbound gRPC deadline so long-lived follows aren't killed
	// by a short client/proxy timeout. Client cancellation still propagates.
	streamCtx, streamCancel := context.WithCancel(context.Background())
	defer streamCancel()
	stop := context.AfterFunc(ctx, streamCancel)
	defer stop()

	logStream, err := s.clientset.CoreV1().Pods(ns).GetLogs(podName, opts).Stream(streamCtx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to stream pod logs: %w", err))
	}
	defer logStream.Close()

	err = podlogs.Stream(ctx, logStream, podlogs.DefaultBatchSize, func(lines []*dataplane.LogLine) error {
		return send(&flyteapp.LogLines{StructuredLines: lines})
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("error reading log stream: %w", err))
	}
	return nil
}

// pickUserContainer returns the primary user container, skipping Knative sidecars.
func pickUserContainer(pod *corev1.Pod) string {
	for _, c := range pod.Spec.Containers {
		if c.Name != knativeQueueProxy {
			return c.Name
		}
	}
	return ""
}
