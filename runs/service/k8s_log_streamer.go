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

	"github.com/flyteorg/flyte/v2/flytestdlib/k8s/podlogs"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

const defaultInitialLines = int64(1000)

// K8sLogStreamer streams logs directly from Kubernetes pods.
type K8sLogStreamer struct {
	clientset kubernetes.Interface
}

// NewK8sLogStreamer creates a K8sLogStreamer from a Kubernetes REST config.
// It clears the timeout so that long-lived log streams are not interrupted.
func NewK8sLogStreamer(k8sConfig *rest.Config) (*K8sLogStreamer, error) {
	cfg := rest.CopyConfig(k8sConfig)
	cfg.Timeout = 0
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	return &K8sLogStreamer{clientset: clientset}, nil
}

// TailLogs streams log lines for the given LogContext from a Kubernetes pod.
func (s *K8sLogStreamer) TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[workflow.TailLogsResponse]) error {
	pod, container, err := getPrimaryPodAndContainer(logContext)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	tailLines := defaultInitialLines
	opts := &corev1.PodLogOptions{
		Container:  container.GetContainerName(),
		Follow:     true,
		Timestamps: true,
		TailLines:  &tailLines,
	}

	// Set SinceTime from container start time if available.
	// When SinceTime is set, it takes precedence and we clear TailLines
	// to stream all logs from that point forward.
	if startTime := container.GetProcess().GetContainerStartTime(); startTime != nil {
		t := metav1.NewTime(startTime.AsTime())
		opts.SinceTime = &t
		opts.TailLines = nil
	}

	// Only follow logs when the pod is actively running. For pending or
	// terminated pods, disable follow so existing logs are returned immediately.
	podObj, err := s.clientset.CoreV1().Pods(pod.GetNamespace()).Get(ctx, pod.GetPodName(), metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("pod %s not found in namespace %s", pod.GetPodName(), pod.GetNamespace()))
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get pod: %w", err))
	}
	opts.Follow = podObj.Status.Phase == corev1.PodRunning

	// Create a context without the incoming gRPC deadline so long-lived follow
	// streams are not killed by a short client/proxy timeout. Cancellation is
	// still propagated so the stream closes when the client disconnects.
	streamCtx, streamCancel := context.WithCancel(context.Background())
	defer streamCancel()
	stop := context.AfterFunc(ctx, streamCancel)
	defer stop()

	logStream, err := s.clientset.CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetPodName(), opts).Stream(streamCtx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to stream pod logs: %w", err))
	}
	defer logStream.Close()

	err = podlogs.Stream(ctx, logStream, podlogs.DefaultBatchSize, func(lines []*dataplane.LogLine) error {
		return stream.Send(&workflow.TailLogsResponse{
			Logs: []*workflow.TailLogsResponse_Logs{{Lines: lines}},
		})
	})
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("error reading log stream: %w", err))
	}
	return nil
}
