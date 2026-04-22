package app

import (
	"context"
	"net/http"

	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"

	appconfig "github.com/flyteorg/flyte/v2/app/config"
	"github.com/flyteorg/flyte/v2/app/service"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// Setup registers the control plane AppService handler on the SetupContext mux.
// In unified mode (sc.BaseURL set), the proxy routes to InternalAppService on
// the same mux via the /internal prefix — no network hop. In split mode,
// cfg.InternalAppServiceURL points at the data plane host.
func Setup(ctx context.Context, sc *stdlibapp.SetupContext, cfg *appconfig.AppConfig) error {
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

	return nil
}
