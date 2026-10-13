package internal

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
	"connectrpc.com/otelconnect"

	appconfig "github.com/flyteorg/flyte/v2/app/internal/config"
	appk8s "github.com/flyteorg/flyte/v2/app/internal/k8s"
	"github.com/flyteorg/flyte/v2/app/internal/service"
	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

const otelServiceName = "internal-app-service"

// Setup registers the InternalAppService handler on the SetupContext mux.
// It is mounted at /internal<path> to avoid collision with the control plane
// AppService, which shares the same proto service definition.
func Setup(ctx context.Context, sc *stdlibapp.SetupContext, cfg *appconfig.InternalAppConfig) error {
	if !cfg.Enabled {
		logger.Infof(ctx, "InternalAppService disabled (apps.enabled=false), skipping setup")
		return nil
	}

	if err := stdlibapp.InitAppScheme(); err != nil {
		return fmt.Errorf("internalapp: failed to register Knative scheme: %w", err)
	}

	appK8sClient := appk8s.NewAppK8sClient(sc.K8sClient, sc.K8sCache, sc.Namespace, cfg)
	internalAppSvc := service.NewInternalAppService(appK8sClient)

	if err := appK8sClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("internalapp: failed to start KService watcher: %w", err)
	}
	sc.AddWorker("app-kservice-watcher", func(ctx context.Context) error {
		<-ctx.Done()
		appK8sClient.StopWatching()
		return nil
	})

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

	handlerServer := connect.NewServer(otelInterceptor)
	appconnect.RegisterAppServiceHandler(handlerServer, internalAppSvc)
	path := "/" + appconnect.AppServiceName + "/"
	handler := http.NewServeMux()
	connecthttp.Mount(handler, handlerServer, connecthttp.WithReadMaxBytes(0))
	sc.Mux.Handle("/internal"+path, http.StripPrefix("/internal", handler))
	logger.Infof(ctx, "Mounted InternalAppService at /internal%s", path)

	if sc.K8sConfig != nil {
		streamer, err := service.NewK8sAppLogStreamer(sc.K8sConfig, sc.Namespace)
		if err != nil {
			return fmt.Errorf("internalapp: failed to create log streamer: %w", err)
		}
		logsSvc := service.NewInternalAppLogsService(appK8sClient, streamer)
		logsHandlerServer := connect.NewServer(otelInterceptor)
		appconnect.RegisterAppLogsServiceHandler(logsHandlerServer, logsSvc)
		logsPath := "/" + appconnect.AppLogsServiceName + "/"
		logsHandler := http.NewServeMux()
		connecthttp.Mount(logsHandler, logsHandlerServer, connecthttp.WithReadMaxBytes(0))
		sc.Mux.Handle("/internal"+logsPath, http.StripPrefix("/internal", logsHandler))
		logger.Infof(ctx, "Mounted InternalAppLogsService at /internal%s", logsPath)
	} else {
		logger.Warnf(ctx, "K8sConfig not set, skipping InternalAppLogsService setup")
	}

	return nil
}
