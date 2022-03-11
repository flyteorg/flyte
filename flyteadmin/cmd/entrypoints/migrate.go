package entrypoints

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.
)

var parentMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command controls migration behavior for the Flyte admin database. Please choose a subcommand.",
}

// This runs all the migrations
var migrateCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run all the migrations for the database",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
		logConfig := logger.GetConfig()

		db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
		if err != nil {
			logger.Fatal(ctx, err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal(ctx, err)
		}

		defer func(deferCtx context.Context) {
			if err = sqlDB.Close(); err != nil {
				logger.Fatal(deferCtx, err)
			}
		}(ctx)

		if err = sqlDB.Ping(); err != nil {
			logger.Fatal(ctx, err)
		}
		m := gormigrate.New(db, gormigrate.DefaultOptions, config.Migrations)
		if err = m.Migrate(); err != nil {
			logger.Fatalf(ctx, "Could not migrate: %v", err)
		}
		logger.Infof(ctx, "Migration ran successfully")
	},
}

// Rollback the latest migration
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "This command will rollback one migration",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
		logConfig := logger.GetConfig()

		db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
		if err != nil {
			logger.Fatal(ctx, err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal(ctx, err)
		}
		defer func(deferCtx context.Context) {
			if err = sqlDB.Close(); err != nil {
				logger.Fatal(deferCtx, err)
			}
		}(ctx)

		if err = sqlDB.Ping(); err != nil {
			logger.Fatal(ctx, err)
		}

		m := gormigrate.New(db, gormigrate.DefaultOptions, config.Migrations)
		err = m.RollbackLast()
		if err != nil {
			logger.Fatalf(ctx, "Could not rollback latest migration: %v", err)
		}
		logger.Infof(ctx, "Rolled back one migration successfully")
	},
}

// This seeds the database with project values
var seedProjectsCmd = &cobra.Command{
	Use:   "seed-projects",
	Short: "Seed projects in the database.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
		logConfig := logger.GetConfig()

		db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
		if err != nil {
			logger.Fatal(ctx, err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal(ctx, err)
		}

		defer func(deferCtx context.Context) {
			if err = sqlDB.Close(); err != nil {
				logger.Fatal(deferCtx, err)
			}
		}(ctx)

		if err = sqlDB.Ping(); err != nil {
			logger.Fatal(ctx, err)
		}

		if err = config.SeedProjects(db, args); err != nil {
			logger.Fatalf(ctx, "Could not add projects to database with err: %v", err)
		}
		logger.Infof(ctx, "Successfully added projects to database")
	},
}

func init() {
	RootCmd.AddCommand(parentMigrateCmd)
	parentMigrateCmd.AddCommand(migrateCmd)
	parentMigrateCmd.AddCommand(rollbackCmd)
	parentMigrateCmd.AddCommand(seedProjectsCmd)
}
