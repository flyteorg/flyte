package cache_service

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/cache_service/migrations"
	"github.com/flyteorg/flyte/v2/cache_service/service"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2/v2connect"
)

const otelServiceName = "cache-service"

// Setup registers the CacheService handler on the SetupContext mux.
// Requires sc.DB and sc.DataStore to be set by the standalone binary.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	cfg := config.GetConfig()
	if err := migrations.RunMigrations(ctx, sc.DB); err != nil {
		return fmt.Errorf("cache_service: failed to run migrations: %w", err)
	}

	otelCfg := otelutils.GetConfig()
	err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg)
	if err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}

	tracerProvider := otelutils.GetTracerProvider(otelServiceName)
	meterProvider := otelutils.GetMeterProvider(otelServiceName)

	otelInterceptor, err := otelconnect.NewInterceptor(otelconnect.WithTracerProvider(tracerProvider), otelconnect.WithMeterProvider(meterProvider), otelconnect.WithoutServerPeerAttributes())
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}

	path, handler := v2connect.NewCacheServiceHandler(service.NewCacheService(cfg, sc.DB), connect.WithInterceptors(otelInterceptor))
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
