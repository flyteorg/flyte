package repositories

import (
	"context"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"reflect"

	errors2 "github.com/flyteorg/flyte/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/runtime"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var migrationsScope = promutils.NewScope("migrations")
var migrateScope = migrationsScope.NewSubScope("migrate")

// all postgres servers come by default with a db name named postgres
const defaultDB = "postgres"

// Migrate This command will run all the migrations for the database
func Migrate(ctx context.Context) error {
	configProvider := runtime.NewConfigurationProvider()
	dbConfigValues := *configProvider.ApplicationConfiguration().GetDbConfig()

	dbName := dbConfigValues.Postgres.DbName
	dbHandle, err := NewDBHandle(ctx, dbConfigValues, migrateScope)

	if err != nil {
		// if db does not exist, try creating it
		_, ok := err.(errors2.ConnectError)
		if !ok {
			logger.Errorf(ctx, "Failed to cast error of type: %v, err: %v", reflect.TypeOf(err),
				err)
			panic(err)
		}

		if database.IsPgErrorWithCode(err, database.PqInvalidDBCode) {
			logger.Warningf(ctx, "Database [%v] does not exist, trying to create it now", dbName)

			dbConfigValues.Postgres.DbName = defaultDB
			setupDBHandler, err := NewDBHandle(ctx, dbConfigValues, migrateScope)
			if err != nil {
				logger.Errorf(ctx, "Failed to connect to default DB %v, err %v", defaultDB, err)
				panic(err)
			}

			// Create the database if it doesn't exist
			// NOTE: this is non-destructive - if for some reason one does exist an err will be thrown
			err = setupDBHandler.CreateDB(dbName)
			if err != nil {
				logger.Errorf(ctx, "Failed to create DB %v err %v", dbName, err)
				panic(err)
			}

			dbConfigValues.Postgres.DbName = dbName
			dbHandle, err = NewDBHandle(ctx, dbConfigValues, migrateScope)
			if err != nil {
				logger.Errorf(ctx, "Failed to connect DB err %v", err)
				panic(err)
			}
		} else {
			logger.Errorf(ctx, "Failed to connect DB err %v", err)
			panic(err)
		}
	}

	logger.Infof(ctx, "Created DB connection.")

	// 	TODO: checkpoints for migrations
	if err := dbHandle.Migrate(ctx); err != nil {
		logger.Errorf(ctx, "Failed to migrate. err: %v", err)
		panic(err)
	}
	logger.Infof(ctx, "Ran DB migration successfully.")
	return nil
}
