package actions

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/flyteorg/flyte/v2/actions/config"
	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	"github.com/flyteorg/flyte/v2/actions/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

const otelServiceName = "actions-service"

// Setup registers the ActionsService handler on the SetupContext mux.
// Requires sc.K8sClient and sc.Namespace to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()

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

	if err := actionsk8s.InitScheme(); err != nil {
		return fmt.Errorf("actions: failed to initialize scheme: %w", err)
	}

	runServiceURL := cfg.RunServiceURL
	if sc.BaseURL != "" {
		runServiceURL = sc.BaseURL
	}
	runClient := workflowconnect.NewInternalRunServiceClient(http.DefaultClient, runServiceURL, connect.WithInterceptors(otelInterceptor))

	actionsClient := actionsk8s.NewActionsClient(
		sc.K8sClient,
		sc.K8sCache,
		cfg.WatchBufferSize,
		cfg.WatchWorkers,
		runClient,
		cfg.RecordFilterSize,
		sc.Scope,
	)
	logger.Infof(ctx, "Actions K8s client initialized")

	if err := actionsClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("actions: failed to start TaskAction watcher: %w", err)
	}
	sc.AddWorker("actions-watcher", func(ctx context.Context) error {
		<-ctx.Done()
		actionsClient.StopWatching()
		return nil
	})

	actionsSvc := service.NewActionsService(actionsClient)

	path, handler := actionsconnect.NewActionsServiceHandler(actionsSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted ActionsService at %s", path)

	return nil
}
