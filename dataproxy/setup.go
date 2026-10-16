package dataproxy

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/dataproxy/logs"
	"github.com/flyteorg/flyte/v2/dataproxy/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/authorization"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster/clusterconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
)

const otelServiceName = "dataproxy-service"

// Setup registers the DataProxy service handler on the SetupContext mux.
// Requires sc.DataStore to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	var authInterceptor *authorization.Interceptor
	if cfg.Authorization.Enabled {
		var err error
		authInterceptor, err = authorization.NewInterceptor(cfg.Authorization, cfg.TrustForwardedIdentityHeaders)
		if err != nil {
			return fmt.Errorf("dataproxy: authorization: %w", err)
		}
		if cfg.RunService.Auth.Type != serviceclient.AuthTypeNone {
			return fmt.Errorf("dataproxy: authorization requires forwarding caller claims; runService service credentials must be disabled")
		}
	}

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

	handlers := []connect.Interceptor{otelInterceptor}
	clients := []connect.Interceptor{otelInterceptor}
	if authInterceptor != nil {
		handlers = append(handlers, authInterceptor)
		clients = append(clients, authorization.ForwardClaims())
	}

	runServiceCfg := cfg.RunService
	if sc.BaseURL != "" {
		runServiceCfg.URL = sc.BaseURL
	}
	runHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, runServiceCfg)
	if err != nil {
		return fmt.Errorf("dataproxy: configure runs service client: %w", err)
	}
	if authInterceptor != nil {
		// Custom claims headers must never follow redirects to another host.
		client := *runHTTPClient
		client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
		runHTTPClient = &client
	}
	taskClient := taskconnect.NewTaskServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clients...))
	triggerClient := triggerconnect.NewTriggerServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clients...))
	runClient := workflowconnect.NewRunServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clients...))
	projectClient := projectconnect.NewProjectServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clients...))

	var logStreamer logs.LogStreamer
	if sc.K8sConfig != nil {
		var err error
		logStreamer, err = logs.NewK8sLogStreamer(sc.K8sConfig)
		if err != nil {
			return fmt.Errorf("failed to create k8s log streamer: %w", err)
		}
	}

	svc := service.NewService(*cfg, sc.DataStore, taskClient, triggerClient, runClient, projectClient, logStreamer)

	path, handler := dataproxyconnect.NewDataProxyServiceHandler(svc, connect.WithInterceptors(handlers...))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", path)

	clusterSvc := service.NewClusterService()
	clusterPath, clusterHandler := clusterconnect.NewClusterServiceHandler(clusterSvc, connect.WithInterceptors(handlers...))
	sc.Mux.Handle(clusterPath, clusterHandler)
	logger.Infof(ctx, "Mounted ClusterService at %s", clusterPath)

	translatorSvc := NewTranslatorService(sc.DataStore, runClient, triggerClient)
	if authInterceptor != nil {
		translatorSvc.EnableAuthorization()
	}
	translatorPath, translatorHandler := workflowconnect.NewTranslatorServiceHandler(translatorSvc, connect.WithInterceptors(handlers...))
	sc.Mux.Handle(translatorPath, translatorHandler)
	logger.Infof(ctx, "Mounted TranslatorService at %s", translatorPath)

	sc.AddReadyCheck(func(r *http.Request) error {
		baseContainer := sc.DataStore.GetBaseContainerFQN(r.Context())
		if baseContainer == "" {
			return fmt.Errorf("storage connection error")
		}
		return nil
	})

	return nil
}
