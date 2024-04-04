package repositories

import (
	"context"

	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/clients/postgres"
	"github.com/flyteorg/flyte/cacheservice/pkg/runtime"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// Migrate This command will run all the migrations for the database
func Migrate(ctx context.Context) error {
	configProvider := runtime.NewConfigurationProvider()

	clientType := configProvider.ApplicationConfiguration().GetCacheServiceConfig().OutputDataStoreType

	switch clientType {
	case "postgres":
		dbConfigValues := configProvider.ApplicationConfiguration().GetDbConfig()
		dbHandle, err := postgres.NewDBHandle(ctx, dbConfigValues)
		if err != nil {
			logger.Errorf(ctx, "Failed to get DB connection. err: %v", err)
			panic(err)
		}

		logger.Infof(ctx, "Created DB connection.")

		if err := dbHandle.Migrate(ctx); err != nil {
			logger.Errorf(ctx, "Failed to migrate. err: %v", err)
			panic(err)
		}
		logger.Infof(ctx, "Ran DB migration successfully.")
	}
	return nil
}
