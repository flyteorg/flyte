package cache_service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/cache_service/migrations"
	"github.com/flyteorg/flyte/v2/cache_service/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	v2connect "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2/v2connect"
)

// Setup registers the CacheService handler on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set by the standalone binary.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	if err := migrations.RunMigrations(ctx, sc.DB); err != nil {
		return fmt.Errorf("cache_service: failed to run migrations: %w", err)
	}

	path, handler := v2connect.NewCacheServiceHandler(service.NewCacheService(cfg, sc.DB))
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted CacheService at %s", path)

	sc.AddReadyCheck(func(r *http.Request) error {
		if err := sc.DB.PingContext(r.Context()); err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}
		if sc.DataStore.GetBaseContainerFQN(r.Context()) == "" {
			return fmt.Errorf("storage connection error")
		}
		return nil
	})

	return nil
}
