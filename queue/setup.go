package queue

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/queue/k8s"
	"github.com/flyteorg/flyte/v2/queue/service"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers the Queue service handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	if err := k8s.InitScheme(); err != nil {
		return fmt.Errorf("queue: failed to initialize scheme: %w", err)
	}

	queueK8sClient := k8s.NewQueueClient(sc.K8sClient, sc.Namespace)
	logger.Infof(ctx, "Queue K8s client initialized for namespace: %s", sc.Namespace)

	queueSvc := service.NewQueueService(queueK8sClient)

	path, handler := workflowconnect.NewQueueServiceHandler(queueSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted QueueService at %s", path)

	return nil
}
