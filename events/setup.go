package events

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/flyteorg/flyte/v2/events/config"
	"github.com/flyteorg/flyte/v2/events/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

const otelServiceName = "events-service"

// Setup registers the EventsProxyService handler.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	runServiceURL := cfg.RunServiceURL
	if sc.BaseURL != "" {
		runServiceURL = sc.BaseURL
	}
	runClient := workflowconnect.NewInternalRunServiceClient(http.DefaultClient, runServiceURL)

	eventsSvc := service.NewEventsProxyService(runClient)

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}

	path, handler := workflowconnect.NewEventsProxyServiceHandler(eventsSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted EventsProxyService at %s", path)

	return nil
}
