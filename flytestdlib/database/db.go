package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// GetDB uses the dbConfig to create a sqlx DB object. If the db doesn't exist for the dbConfig then a new one is created
// using the default db for the provider. eg : postgres has default dbName as postgres
func GetDB(ctx context.Context, dbConfig *DbConfig) (*sqlx.DB, error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}

	var db *sqlx.DB
	var err error

	switch {
	case !(dbConfig.Postgres.IsEmpty()):
		db, err = CreatePostgresDbIfNotExists(ctx, dbConfig.Postgres)
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
		db, err = CreatePostgresDbIfNotExists(ctx, pgConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unrecognized database config, %v. Postgres config is required", dbConfig)
	}

	// Setup connection pool settings
	setupDbConnectionPool(ctx, db, dbConfig)
	return db, nil
}

// GetReadOnlyDB uses the dbConfig to create a sqlx DB object for the read replica passed via the config
func GetReadOnlyDB(ctx context.Context, dbConfig *DbConfig) (*sqlx.DB, error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}

	if dbConfig.Postgres.IsEmpty() || dbConfig.Postgres.ReadReplicaHost == "" {
		return nil, fmt.Errorf("read replica host not provided in db config")
	}

	db, err := CreatePostgresReadOnlyDbConnection(ctx, dbConfig.Postgres)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupDbConnectionPool(ctx context.Context, db *sqlx.DB, dbConfig *DbConfig) {
	db.SetConnMaxLifetime(dbConfig.ConnMaxLifeTime.Duration)
	db.SetMaxIdleConns(dbConfig.MaxIdleConnections)
	db.SetMaxOpenConns(dbConfig.MaxOpenConnections)
	logger.Infof(ctx, "Set connection pool values to [%+v]", db.Stats())
}
