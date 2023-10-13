package database

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const pqInvalidDBCode = "3D000"
const pqDbAlreadyExistsCode = "42P04"

// GetDB uses the dbConfig to create gorm DB object. If the db doesn't exist for the dbConfig then a new one is created
// using the default db for the provider. eg : postgres has default dbName as postgres
func GetDB(ctx context.Context, dbConfig *DbConfig, logConfig *logger.Config) (
	*gorm.DB, error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}
	gormConfig := &gorm.Config{
		Logger:                                   database.GetGormLogger(ctx, logConfig),
		DisableForeignKeyConstraintWhenMigrating: !dbConfig.EnableForeignKeyConstraintWhenMigrating,
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

func setupDbConnectionPool(ctx context.Context, gormDb *gorm.DB, dbConfig *database.DbConfig) error {
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
