package app

import (
	"context"
	"net/http"

	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"

	appconfig "github.com/flyteorg/flyte/v2/app/config"
	appinternal "github.com/flyteorg/flyte/v2/app/internal"
	"github.com/flyteorg/flyte/v2/app/service"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// SetupInternal registers the data plane InternalAppService on the SetupContext mux.
// It must be called before Setup so the proxy can reach /internal/... on the same mux.
func SetupInternal(ctx context.Context, sc *stdlibapp.SetupContext) error {
	return appinternal.Setup(ctx, sc, appconfig.GetInternalAppConfig())
}

// Setup registers the control plane AppService handler on the SetupContext mux.
// In unified mode (sc.BaseURL set), the proxy routes to InternalAppService on
// the same mux via the /internal prefix — no network hop. In split mode,
// cfg.InternalAppServiceURL points at the data plane host.
func Setup(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := appconfig.GetAppConfig()

	internalAppURL := cfg.InternalAppServiceURL
	if sc.BaseURL != "" {
		internalAppURL = sc.BaseURL
	}

	internalClient := appconnect.NewAppServiceClient(
		http.DefaultClient,
		internalAppURL+"/internal",
	)

	appSvc := service.NewAppService(internalClient, cfg.CacheTTL)
	path, handler := appconnect.NewAppServiceHandler(appSvc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted AppService at %s", path)

	internalLogsClient := appconnect.NewAppLogsServiceClient(
		http.DefaultClient,
		internalAppURL+"/internal",
	)
	logsSvc := service.NewAppLogsService(internalLogsClient)
	logsPath, logsHandler := appconnect.NewAppLogsServiceHandler(logsSvc)
	sc.Mux.Handle(logsPath, logsHandler)
	logger.Infof(ctx, "Mounted AppLogsService at %s", logsPath)

	return nil
}
