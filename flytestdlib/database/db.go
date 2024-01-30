package database

import (
	"context"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// GetDB uses the dbConfig to create gorm DB object. If the db doesn't exist for the dbConfig then a new one is created
// using the default db for the provider. eg : postgres has default dbName as postgres
func GetDB(ctx context.Context, dbConfig *DbConfig, logConfig *logger.Config) (
	*gorm.DB, error) {

	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}
	gormConfig := &gorm.Config{
		Logger:                                   GetGormLogger(ctx, logConfig),
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	var gormDb *gorm.DB
	var err error

	switch {
	case !(dbConfig.SQLite.IsEmpty()):
		if dbConfig.SQLite.File == "" {
			return nil, fmt.Errorf("illegal sqlite database configuration. `file` is a required parameter and should be a path")
		}
		gormDb, err = gorm.Open(sqlite.Open(dbConfig.SQLite.File), gormConfig)
		if err != nil {
			return nil, err
		}
	case !(dbConfig.Postgres.IsEmpty()):
		gormDb, err = CreatePostgresDbIfNotExists(ctx, gormConfig, dbConfig.Postgres)
		if err != nil {
			return nil, err
		}

	case len(dbConfig.DeprecatedHost) > 0 || len(dbConfig.DeprecatedUser) > 0 || len(dbConfig.DeprecatedDbName) > 0:
		pgConfig := PostgresConfig{
			Host:         dbConfig.DeprecatedHost,
			Port:         dbConfig.DeprecatedPort,
			DbName:       dbConfig.DeprecatedDbName,
			User:         dbConfig.DeprecatedUser,
			Password:     dbConfig.DeprecatedPassword,
			PasswordPath: dbConfig.DeprecatedPasswordPath,
			ExtraOptions: dbConfig.DeprecatedExtraOptions,
			Debug:        dbConfig.DeprecatedDebug,
		}
		gormDb, err = CreatePostgresDbIfNotExists(ctx, gormConfig, pgConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unrecognized database config, %v. Supported only postgres and sqlite", dbConfig)
	}

	// Setup connection pool settings
	return gormDb, setupDbConnectionPool(ctx, gormDb, dbConfig)
}

func setupDbConnectionPool(ctx context.Context, gormDb *gorm.DB, dbConfig *DbConfig) error {
	genericDb, err := gormDb.DB()
	if err != nil {
		return err
	}
	genericDb.SetConnMaxLifetime(dbConfig.ConnMaxLifeTime.Duration)
	genericDb.SetMaxIdleConns(dbConfig.MaxIdleConnections)
	genericDb.SetMaxOpenConns(dbConfig.MaxOpenConnections)
	logger.Infof(ctx, "Set connection pool values to [%+v]", genericDb.Stats())
	return nil
}

func withDB(ctx context.Context, databaseConfig *DbConfig, do func(db *gorm.DB) error) error {
	if databaseConfig == nil {
		databaseConfig = GetConfig()
	}
	logConfig := logger.GetConfig()

	db, err := GetDB(ctx, databaseConfig, logConfig)
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
func Migrate(ctx context.Context, databaseConfig *DbConfig, migrations []*gormigrate.Migration) error {
	if len(migrations) == 0 {
		logger.Infof(ctx, "No migrations to run")
		return nil
	}
	return withDB(ctx, databaseConfig, func(db *gorm.DB) error {
		m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
		if err := m.Migrate(); err != nil {
			return fmt.Errorf("database migration failed: %v", err)
		}
		logger.Infof(ctx, "Migration ran successfully")
		return nil
	})
}

// Rollback rolls back the last migration
func Rollback(ctx context.Context, databaseConfig *DbConfig, migrations []*gormigrate.Migration) error {
	if len(migrations) == 0 {
		logger.Infof(ctx, "No migrations to rollback")
		return nil
	}
	return withDB(ctx, databaseConfig, func(db *gorm.DB) error {
		m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
		err := m.RollbackLast()
		if err != nil {
			return fmt.Errorf("could not rollback latest migration: %v", err)
		}
		logger.Infof(ctx, "Rolled back one migration successfully")
		return nil
	})
}
