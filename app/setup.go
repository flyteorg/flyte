package app

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/sentryutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"

	appconfig "github.com/flyteorg/flyte/v2/app/config"
	appinternal "github.com/flyteorg/flyte/v2/app/internal"
	"github.com/flyteorg/flyte/v2/app/service"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

const otelServiceName = "app-service"

var sentryOperations = map[string]string{
	appconnect.AppServiceCreateProcedure: "deploy_app",
}

// SetupInternal registers the data plane InternalAppService on the SetupContext mux.
// It must be called before Setup so the proxy can reach /internal/... on the same mux.
func SetupInternal(ctx context.Context, sc *stdlibapp.SetupContext) error {
	return appinternal.Setup(ctx, sc, appconfig.GetInternalAppConfig())
}

// Setup registers the control plane AppService handler on the SetupContext mux.
// In unified mode (sc.BaseURL set), the proxy routes to InternalAppService on
// the same mux via the /internal prefix — no network hop. In split mode,
// cfg.InternalAppService points at the data plane host.
func Setup(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := appconfig.GetAppConfig()

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
		otelconnect.WithoutServerPeerAttributes(),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}

	interceptors := []connect.Interceptor{otelInterceptor}
	if sentryutils.Init(ctx, otelServiceName) {
		interceptors = append(interceptors, sentryutils.Interceptor(sentryOperations))
		sc.AddWorker("app-sentry-flush", func(ctx context.Context) error {
			<-ctx.Done()
			sentryutils.Flush()
			return nil
		})
	}

	internalAppServiceCfg := cfg.InternalAppService
	if sc.BaseURL != "" {
		internalAppServiceCfg.URL = sc.BaseURL
	}
	internalHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, internalAppServiceCfg)
	if err != nil {
		return fmt.Errorf("app: configure internal app service client: %w", err)
	}
	internalAppURL := internalAppServiceCfg.URL + "/internal"

	internalClient := appconnect.NewAppServiceClient(
		internalHTTPClient,
		internalAppURL,
		connect.WithInterceptors(otelInterceptor),
	)

	appSvc := service.NewAppService(internalClient, cfg.CacheTTL)

	path, handler := appconnect.NewAppServiceHandler(appSvc, connect.WithInterceptors(interceptors...))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted AppService at %s", path)

	internalLogsClient := appconnect.NewAppLogsServiceClient(
		internalHTTPClient,
		internalAppURL,
		connect.WithInterceptors(otelInterceptor),
	)
	logsSvc := service.NewAppLogsService(internalLogsClient)
	logsPath, logsHandler := appconnect.NewAppLogsServiceHandler(logsSvc, connect.WithInterceptors(otelInterceptor))
	sc.Mux.Handle(logsPath, logsHandler)
	logger.Infof(ctx, "Mounted AppLogsService at %s", logsPath)

	return nil
}
