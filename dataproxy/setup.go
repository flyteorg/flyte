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
	flyteauth "github.com/flyteorg/flyte/v2/flytestdlib/authorization/flyte"
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
		authInterceptor, err = authorization.New(cfg.Authorization, flyteauth.Resolve)
		if err != nil {
			return fmt.Errorf("dataproxy authorization: %w", err)
		}
		if cfg.RunService.Auth.Type != serviceclient.AuthTypeNone {
			return fmt.Errorf("dataproxy Cedar requires forwarding caller claims; Runs service credentials must be disabled")
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

	serverInterceptors := []connect.Interceptor{otelInterceptor}
	clientInterceptors := []connect.Interceptor{otelInterceptor}
	if authInterceptor != nil {
		serverInterceptors = append(serverInterceptors, authInterceptor)
		clientInterceptors = append(clientInterceptors, authorization.ForwardClaims())
	}

	runServiceCfg := cfg.RunService
	if sc.BaseURL != "" {
		runServiceCfg.URL = sc.BaseURL
	}
	runHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, runServiceCfg)
	if err != nil {
		return fmt.Errorf("dataproxy: configure runs service client: %w", err)
	}
	taskClient := taskconnect.NewTaskServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clientInterceptors...))
	triggerClient := triggerconnect.NewTriggerServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clientInterceptors...))
	runClient := workflowconnect.NewRunServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clientInterceptors...))
	projectClient := projectconnect.NewProjectServiceClient(runHTTPClient, runServiceCfg.URL, connect.WithInterceptors(clientInterceptors...))

	var logStreamer logs.LogStreamer
	if sc.K8sConfig != nil {
		var err error
		logStreamer, err = logs.NewK8sLogStreamer(sc.K8sConfig)
		if err != nil {
			return fmt.Errorf("failed to create k8s log streamer: %w", err)
		}
	}

	svc := service.NewService(*cfg, sc.DataStore, taskClient, triggerClient, runClient, projectClient, logStreamer)

	path, handler := dataproxyconnect.NewDataProxyServiceHandler(svc, connect.WithInterceptors(serverInterceptors...))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", path)

	clusterSvc := service.NewClusterService()
	clusterPath, clusterHandler := clusterconnect.NewClusterServiceHandler(clusterSvc, connect.WithInterceptors(serverInterceptors...))
	sc.Mux.Handle(clusterPath, clusterHandler)
	logger.Infof(ctx, "Mounted ClusterService at %s", clusterPath)

	translatorSvc := NewTranslatorService(sc.DataStore, runClient, triggerClient)
	translatorPath, translatorHandler := workflowconnect.NewTranslatorServiceHandler(translatorSvc, connect.WithInterceptors(serverInterceptors...))
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
