package events

import (
	"context"
	"net/http"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/events/config"
	"github.com/flyteorg/flyte/v2/events/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// Setup registers the EventsService handler.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	runServiceURL := cfg.RunServiceURL
	if sc.BaseURL != "" {
		runServiceURL = sc.BaseURL
	}
	runClient := workflowconnect.NewInternalRunServiceClient(http.DefaultClient, runServiceURL)

	eventsSvc := service.NewEventService(runClient)
	path, handler := workflowconnect.NewEventsServiceHandler(eventsSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted EventsService at %s", path)

	return nil
}
