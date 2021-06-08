package entrypoints

import (
	"reflect"

	"github.com/flyteorg/datacatalog/pkg/repositories"
	errors2 "github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/jackc/pgconn"

	"context"

	"github.com/spf13/cobra"
)

var parentMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command controls migration behavior for the Flyte Catalog database. Please choose a subcommand.",
}

var migrationsScope = promutils.NewScope("migrations")
var migrateScope = migrationsScope.NewSubScope("migrate")

// all postgres servers come by default with a db name named postgres
const defaultDB = "postgres"
const pqInvalidDBCode = "3D000"

// This runs all the migrations
var migrateCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run all the migrations for the database",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configProvider := runtime.NewConfigurationProvider()
		dbConfigValues := configProvider.ApplicationConfiguration().GetDbConfig()

		dbName := dbConfigValues.DbName
		dbHandle, err := repositories.NewDBHandle(dbConfigValues, migrateScope)

		if err != nil {
			// if db does not exist, try creating it
			cErr, ok := err.(errors2.ConnectError)
			if !ok {
				logger.Errorf(ctx, "Failed to cast error of type: %v, err: %v", reflect.TypeOf(err),
					err)
				panic(err)
			}
			pqError := cErr.Unwrap().(*pgconn.PgError)
			if pqError.Code == pqInvalidDBCode {
				logger.Warningf(ctx, "Database [%v] does not exist, trying to create it now", dbName)

				dbConfigValues.DbName = defaultDB
				setupDBHandler, err := repositories.NewDBHandle(dbConfigValues, migrateScope)
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

				dbConfigValues.DbName = dbName
				dbHandle, err = repositories.NewDBHandle(dbConfigValues, migrateScope)
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
	},
}

func init() {
	RootCmd.AddCommand(parentMigrateCmd)
	parentMigrateCmd.AddCommand(migrateCmd)
}
