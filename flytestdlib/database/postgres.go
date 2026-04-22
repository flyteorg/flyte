package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	oldPgConn "github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const PqInvalidDBCode = "3D000"
const PqDbAlreadyExistsCode = "42P04"
const PgDuplicatedForeignKey = "23503"
const PgDuplicatedKey = "23505"
const defaultDB = "postgres"

// Resolves a password value from either a user-provided inline value or a filepath whose contents contain a password.
func resolvePassword(ctx context.Context, passwordVal, passwordPath string) string {
	password := passwordVal
	if len(passwordPath) > 0 {
		if _, err := os.Stat(passwordPath); os.IsNotExist(err) {
			logger.Fatalf(ctx,
				"missing database password at specified path [%s]", passwordPath)
		}
		passwordVal, err := os.ReadFile(passwordPath)
		if err != nil {
			logger.Fatalf(ctx, "failed to read database password from path [%s] with err: %v",
				passwordPath, err)
		}
		// Passwords can contain special characters as long as they are percent encoded
		// https://www.postgresql.org/docs/current/libpq-connect.html
		password = strings.TrimSpace(string(passwordVal))
	}
	return password
}

// Produces the DSN (data source name) for opening a postgres db connection.
func GetPostgresDsn(ctx context.Context, pgConfig PostgresConfig) string {
	password := resolvePassword(ctx, pgConfig.Password, pgConfig.PasswordPath)
	if len(password) == 0 {
		// The password-less case is included for development environments.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User, password, pgConfig.ExtraOptions)
}

// Produces the DSN (data source name) for the read replica for opening a postgres db connection.
func getPostgresReadDsn(ctx context.Context, pgConfig PostgresConfig) string {
	password := resolvePassword(ctx, pgConfig.Password, pgConfig.PasswordPath)
	if len(password) == 0 {
		// The password-less case is included for development environments.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			pgConfig.ReadReplicaHost, pgConfig.Port, pgConfig.DbName, pgConfig.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		pgConfig.ReadReplicaHost, pgConfig.Port, pgConfig.DbName, pgConfig.User, password, pgConfig.ExtraOptions)
}

// CreatePostgresDbIfNotExists creates DB if it doesn't exist for the passed in config
func CreatePostgresDbIfNotExists(ctx context.Context, pgConfig PostgresConfig) (*sqlx.DB, error) {
	dsn := GetPostgresDsn(ctx, pgConfig)
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// sqlx.Open doesn't actually connect; we need to ping to verify the connection.
	if err = db.PingContext(ctx); err == nil {
		return db, nil
	}

	if !IsPgErrorWithCode(err, PqInvalidDBCode) {
		logger.Errorf(ctx, "Unhandled error connecting to postgres, pg [%v]: %v", pgConfig, err)
		db.Close()
		return nil, err
	}
	db.Close()
	logger.Warningf(ctx, "Database [%v] does not exist", pgConfig.DbName)

	// Every postgres installation includes a 'postgres' database by default. We connect to that now in order to
	// initialize the user-specified database.
	defaultDbPgConfig := pgConfig
	defaultDbPgConfig.DbName = defaultDB
	defaultDsn := GetPostgresDsn(ctx, defaultDbPgConfig)
	defaultDb, err := sqlx.Open("pgx", defaultDsn)
	if err != nil {
		return nil, err
	}
	defer defaultDb.Close()

	if err = defaultDb.PingContext(ctx); err != nil {
		return nil, err
	}

	// Because we asserted earlier that the db does not exist, we create it now.
	logger.Infof(ctx, "Creating database %v", pgConfig.DbName)

	// NOTE: golang sql drivers do not support parameter injection for CREATE calls
	createDBStatement := fmt.Sprintf("CREATE DATABASE %s", pgConfig.DbName)
	_, err = defaultDb.ExecContext(ctx, createDBStatement)

	if err != nil {
		if !IsPgErrorWithCode(err, PqDbAlreadyExistsCode) {
			return nil, err
		}
		logger.Warningf(ctx, "Got DB already exists error for [%s], skipping...", pgConfig.DbName)
	}

	// Now try connecting to the db again
	db, err = sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// CreatePostgresReadOnlyDbConnection creates readonly DB connection and returns the sqlx.DB object and error
func CreatePostgresReadOnlyDbConnection(ctx context.Context, pgConfig PostgresConfig) (*sqlx.DB, error) {
	dsn := getPostgresReadDsn(ctx, pgConfig)
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func IsPgErrorWithCode(err error, code string) bool {
	// Newer versions of the gorm postgres driver seem to use
	// "github.com/jackc/pgx/v5/pgconn"
	// See https://github.com/go-gorm/gorm/issues/4135
	// Let's just try both of them to make sure.
	pgErr := &pgconn.PgError{}
	if errors.As(err, &pgErr) {
		// err chain does not contain a pgconn.PgError
		return pgErr.Code == code
	}

	oldPgErr := &oldPgConn.PgError{}
	if errors.As(err, &oldPgErr) {
		// err chain does not contain a pgconn.PgError
		return oldPgErr.Code == code
	}

	return false
}
