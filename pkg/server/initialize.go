package server

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func withDB(ctx context.Context, do func(db *gorm.DB) error) error {
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
		return err
	}

	return do(db)
}

// Migrate runs all configured migrations
func Migrate(ctx context.Context) error {
	return withDB(ctx, func(db *gorm.DB) error {
		m := gormigrate.New(db, gormigrate.DefaultOptions, config.Migrations)
		if err := m.Migrate(); err != nil {
			return fmt.Errorf("database migration failed: %v", err)
		}
		logger.Infof(ctx, "Migration ran successfully")
		return nil
	})
}

// Rollback rolls back the last migration
func Rollback(ctx context.Context) error {
	return withDB(ctx, func(db *gorm.DB) error {
		m := gormigrate.New(db, gormigrate.DefaultOptions, config.Migrations)
		err := m.RollbackLast()
		if err != nil {
			return fmt.Errorf("could not rollback latest migration: %v", err)
		}
		logger.Infof(ctx, "Rolled back one migration successfully")
		return nil
	})
}

// SeedProjects creates a set of given projects in the DB
func SeedProjects(ctx context.Context, projects []string) error {
	return withDB(ctx, func(db *gorm.DB) error {
		if err := config.SeedProjects(db, projects); err != nil {
			return fmt.Errorf("could not add projects to database with err: %v", err)
		}
		logger.Infof(ctx, "Successfully added projects to database")
		return nil
	})
}
