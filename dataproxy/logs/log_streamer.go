package logs

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/samber/lo"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

// LogStreamer abstracts log fetching from different backends.
type LogStreamer interface {
	TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[dataproxy.TailLogsResponse]) error
}

// GetPrimaryPodAndContainer finds the primary pod and container from a LogContext.
func GetPrimaryPodAndContainer(logContext *core.LogContext) (*core.PodLogContext, *core.ContainerContext, error) {
	if logContext.GetPrimaryPodName() == "" {
		return nil, nil, fmt.Errorf("primary pod name is empty in log context")
	}

	pod, found := lo.Find(logContext.GetPods(), func(pod *core.PodLogContext) bool {
		return pod.GetPodName() == logContext.GetPrimaryPodName()
	})
	if !found {
		return nil, nil, fmt.Errorf("primary pod %s not found in log context", logContext.GetPrimaryPodName())
	}

	container, found := lo.Find(pod.GetContainers(), func(c *core.ContainerContext) bool {
		return c.GetContainerName() == pod.GetPrimaryContainerName()
	})
	if !found {
		return nil, nil, fmt.Errorf("primary container %s not found in pod %s", pod.GetPrimaryContainerName(), pod.GetPodName())
	}

	return pod, container, nil
}
