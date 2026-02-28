package dataproxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/dataproxy/service"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Setup registers the DataProxy service handler on the SetupContext mux.
// Requires sc.DataStore to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	svc := service.NewService(*cfg, sc.DataStore)

	path, handler := dataproxyconnect.NewDataProxyServiceHandler(svc)
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", path)

	sc.AddReadyCheck(func(r *http.Request) error {
		baseContainer := sc.DataStore.GetBaseContainerFQN(r.Context())
		if baseContainer == "" {
			return fmt.Errorf("storage connection error")
		}
		return nil
	})

	return nil
}
