package cache_service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/cache_service/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	v2connect "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2/v2connect"
)

// Setup registers the CacheService handler on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set by the standalone binary.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	_ = config.GetConfig()

	path, handler := v2connect.NewCacheServiceHandler(service.NewCacheService())
	sc.Mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted CacheService at %s", path)

	sc.AddReadyCheck(func(r *http.Request) error {
		sqlDB, err := sc.DB.DB()
		if err != nil {
			return fmt.Errorf("database connection error: %w", err)
		}
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("database ping failed: %w", err)
		}
		if sc.DataStore.GetBaseContainerFQN(r.Context()) == "" {
			return fmt.Errorf("storage connection error")
		}
		return nil
	})

	return nil
}
