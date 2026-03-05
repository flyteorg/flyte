package events

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/events/config"
	"github.com/flyteorg/flyte/v2/events/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// Setup registers the EventsService handler and starts background workers.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	runServiceURL := cfg.RunServiceURL
	if sc.BaseURL != "" {
		runServiceURL = sc.BaseURL
	}
	runClient := workflowconnect.NewInternalRunServiceClient(http.DefaultClient, runServiceURL)

	// Create and start events worker.
	eventWorker, err := service.NewEventsWorker(runClient, cfg.QueueSize, cfg.WorkerCount)
	if err != nil {
		return fmt.Errorf("events: failed to create worker: %w", err)
	}
	if err := eventWorker.Start(ctx); err != nil {
		return fmt.Errorf("events: failed to start worker: %w", err)
	}
	sc.AddWorker("events-worker", func(ctx context.Context) error {
		<-ctx.Done()
		eventWorker.End()
		return nil
	})
	logger.Infof(ctx, "Events worker started with queue size %d and %d workers", cfg.QueueSize, cfg.WorkerCount)

	eventsSvc := service.NewEventService(eventWorker)
	path, handler := workflowconnect.NewEventsServiceHandler(eventsSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted EventsService at %s", path)

	return nil
}
