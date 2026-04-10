package dataproxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/dataproxy/service"
	runsService "github.com/flyteorg/flyte/v2/runs/service"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers the DataProxy service handler on the SetupContext mux.
// Requires sc.DataStore to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	baseURL := sc.BaseURL
	taskClient := taskconnect.NewTaskServiceClient(http.DefaultClient, baseURL)
	triggerClient := triggerconnect.NewTriggerServiceClient(http.DefaultClient, baseURL)
	runClient := workflowconnect.NewRunServiceClient(http.DefaultClient, baseURL)

	var logStreamer service.LogStreamer
	if sc.K8sConfig != nil {
		var err error
		logStreamer, err = runsService.NewK8sLogStreamer(sc.K8sConfig)
		if err != nil {
			return fmt.Errorf("failed to create k8s log streamer: %w", err)
		}
	}

	svc := service.NewService(*cfg, sc.DataStore, taskClient, triggerClient, runClient, logStreamer)

	path, handler := dataproxyconnect.NewDataProxyServiceHandler(svc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", path)

	sc.AddReadyCheck(func(r *http.Request) error {
		baseContainer := sc.DataStore.GetBaseContainerFQN(r.Context())
		if baseContainer == "" {
			return fmt.Errorf("storage connection error")
		}
		return nil
	})

	return nil
}
