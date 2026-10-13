package events

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/otelconnect"

	"github.com/flyteorg/flyte/v2/events/config"
	"github.com/flyteorg/flyte/v2/events/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

const otelServiceName = "events-service"

// Setup registers the EventsProxyService handler.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	otelInterceptor, err := otelconnect.NewServerInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}
	otelClientInterceptor, err := otelconnect.NewClientInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
	)
	if err != nil {
		return fmt.Errorf("creating otel client interceptor: %w", err)
	}

	runServiceCfg := cfg.RunService
	if sc.BaseURL != "" {
		runServiceCfg.URL = sc.BaseURL
	}
	runHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, runServiceCfg)
	if err != nil {
		return fmt.Errorf("events: configure run service client: %w", err)
	}
	runClient := workflowconnect.NewInternalRunServiceClient(connect.NewClient(connecthttp.NewTransport(runHTTPClient, runServiceCfg.URL, connecthttp.WithReadMaxBytes(0)), otelClientInterceptor))

	eventsSvc := service.NewEventsProxyService(runClient)

	handlerServer := connect.NewServer(otelInterceptor)
	workflowconnect.RegisterEventsProxyServiceHandler(handlerServer, eventsSvc)
	path := "/" + workflowconnect.EventsProxyServiceName + "/"
	handler := http.NewServeMux()
	connecthttp.Mount(handler, handlerServer, connecthttp.WithReadMaxBytes(0))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted EventsProxyService at %s", path)

	return nil
}
