package dataproxy

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/otelconnect"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/dataproxy/logs"
	"github.com/flyteorg/flyte/v2/dataproxy/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster/clusterconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

const otelServiceName = "dataproxy-service"

// Setup registers the DataProxy service handler on the SetupContext mux.
// Requires sc.DataStore to be set.
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
		return fmt.Errorf("dataproxy: configure runs service client: %w", err)
	}
	taskClient := taskconnect.NewTaskServiceClient(connect.NewClient(connecthttp.NewTransport(runHTTPClient, runServiceCfg.URL, connecthttp.WithReadMaxBytes(0)), otelClientInterceptor))
	triggerClient := triggerconnect.NewTriggerServiceClient(connect.NewClient(connecthttp.NewTransport(runHTTPClient, runServiceCfg.URL, connecthttp.WithReadMaxBytes(0)), otelClientInterceptor))
	runClient := workflowconnect.NewRunServiceClient(connect.NewClient(connecthttp.NewTransport(runHTTPClient, runServiceCfg.URL, connecthttp.WithReadMaxBytes(0)), otelClientInterceptor))
	projectClient := projectconnect.NewProjectServiceClient(connect.NewClient(connecthttp.NewTransport(runHTTPClient, runServiceCfg.URL, connecthttp.WithReadMaxBytes(0)), otelClientInterceptor))

	var logStreamer logs.LogStreamer
	if sc.K8sConfig != nil {
		var err error
		logStreamer, err = logs.NewK8sLogStreamer(sc.K8sConfig)
		if err != nil {
			return fmt.Errorf("failed to create k8s log streamer: %w", err)
		}
	}

	svc := service.NewService(*cfg, sc.DataStore, taskClient, triggerClient, runClient, projectClient, logStreamer)

	handlerServer := connect.NewServer(otelInterceptor)
	dataproxyconnect.RegisterDataProxyServiceHandler(handlerServer, svc)
	path := "/" + dataproxyconnect.DataProxyServiceName + "/"
	handler := http.NewServeMux()
	connecthttp.Mount(handler, handlerServer, connecthttp.WithReadMaxBytes(0))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", path)

	clusterSvc := service.NewClusterService()
	clusterHandlerServer := connect.NewServer(otelInterceptor)
	clusterconnect.RegisterClusterServiceHandler(clusterHandlerServer, clusterSvc)
	clusterPath := "/" + clusterconnect.ClusterServiceName + "/"
	clusterHandler := http.NewServeMux()
	connecthttp.Mount(clusterHandler, clusterHandlerServer, connecthttp.WithReadMaxBytes(0))
	sc.Mux.Handle(clusterPath, clusterHandler)
	logger.Infof(ctx, "Mounted ClusterService at %s", clusterPath)

	translatorSvc := NewTranslatorService(sc.DataStore, runClient, triggerClient)
	translatorHandlerServer := connect.NewServer(otelInterceptor)
	workflowconnect.RegisterTranslatorServiceHandler(translatorHandlerServer, translatorSvc)
	translatorPath := "/" + workflowconnect.TranslatorServiceName + "/"
	translatorHandler := http.NewServeMux()
	connecthttp.Mount(translatorHandler, translatorHandlerServer, connecthttp.WithReadMaxBytes(0))
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
