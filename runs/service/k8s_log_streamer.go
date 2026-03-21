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

const logBatchSize = 100

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

	opts := &corev1.PodLogOptions{
		Container:  container.GetContainerName(),
		Follow:     true,
		Timestamps: true,
	}

	// Set SinceTime from container start time if available.
	if startTime := container.GetProcess().GetContainerStartTime(); startTime != nil {
		t := metav1.NewTime(startTime.AsTime())
		opts.SinceTime = &t
	}

	logStream, err := s.clientset.CoreV1().Pods(pod.GetNamespace()).GetLogs(pod.GetPodName(), opts).Stream(ctx)
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to stream pod logs: %w", err))
	}
	defer logStream.Close()

	scanner := bufio.NewScanner(logStream)
	var lines []*dataplane.LogLine

	for scanner.Scan() {
		line := scanner.Text()
		logLine := parseLogLine(line)
		lines = append(lines, logLine)

		if len(lines) >= logBatchSize {
			if err := stream.Send(&workflow.TailLogsResponse{
				Logs: []*workflow.TailLogsResponse_Logs{
					{Lines: lines},
				},
			}); err != nil {
				return err
			}
			lines = nil
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

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return connect.NewError(connect.CodeInternal, fmt.Errorf("error reading log stream: %w", err))
	}

	return nil
}

// parseLogLine splits a K8s log line into timestamp and message.
// K8s log lines with timestamps are formatted as: "2006-01-02T15:04:05.999999999Z message"
func parseLogLine(line string) *dataplane.LogLine {
	logLine := &dataplane.LogLine{
		Originator: dataplane.LogLineOriginator_USER,
	}

	parts := strings.SplitN(line, " ", 2)
	if len(parts) == 2 {
		if t, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
			logLine.Timestamp = timestamppb.New(t)
			logLine.Message = parts[1]
			return logLine
		}
	}

	// Fallback: entire line is the message.
	logLine.Message = line
	return logLine
}
