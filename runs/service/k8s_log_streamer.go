package service

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

const (
	logBatchSize        = 100
	defaultInitialLines = int64(1000)
)

// K8sLogStreamer streams logs directly from Kubernetes pods.
type K8sLogStreamer struct {
	clientset kubernetes.Interface
}

// NewK8sLogStreamer creates a K8sLogStreamer from a Kubernetes REST config.
func NewK8sLogStreamer(k8sConfig *rest.Config) (*K8sLogStreamer, error) {
	clientset, err := kubernetes.NewForConfig(k8sConfig)
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

	logStream, err := s.clientset.CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetPodName(), opts).Stream(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to stream pod logs: %w", err))
	}
	defer logStream.Close()

	reader := bufio.NewReader(logStream)

	lines := make([]*dataplane.LogLine, 0, logBatchSize)

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// Trim trailing newline.
			line = strings.TrimRight(line, "\n")
			logLine := parseLogLine(line)
			lines = append(lines, logLine)

			if len(lines) >= logBatchSize {
				if sendErr := stream.Send(&workflow.TailLogsResponse{
					Logs: []*workflow.TailLogsResponse_Logs{
						{Lines: lines},
					},
				}); sendErr != nil {
					return sendErr
				}
				lines = make([]*dataplane.LogLine, 0, logBatchSize)
			}
		}
		if err != nil {
			break
		}
	}

	// Send remaining lines.
	if len(lines) > 0 {
		if err := stream.Send(&workflow.TailLogsResponse{
			Logs: []*workflow.TailLogsResponse_Logs{
				{Lines: lines},
			},
		}); err != nil {
			return err
		}
	}

	if ctx.Err() != nil {
		return nil
	}

	return nil
}

// parseLogLine splits a K8s log line into timestamp and message.
// K8s log lines with timestamps are formatted as: "2006-01-02T15:04:05.999999999Z message"
func parseLogLine(line string) *dataplane.LogLine {
	if idx := strings.IndexByte(line, ' '); idx > 0 {
		if t, err := time.Parse(time.RFC3339Nano, line[:idx]); err == nil {
			return &dataplane.LogLine{
				Originator: dataplane.LogLineOriginator_USER,
				Timestamp:  timestamppb.New(t),
				Message:    line[idx+1:],
			}
		}
	}

	return &dataplane.LogLine{
		Originator: dataplane.LogLineOriginator_USER,
		Message:    line,
	}
}
