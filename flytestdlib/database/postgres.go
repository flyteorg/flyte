package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

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
// With IAM auth the DSN carries no password; a token is attached to each new connection instead.
func GetPostgresDsn(ctx context.Context, pgConfig PostgresConfig) string {
	return postgresDsn(ctx, pgConfig, pgConfig.Host)
}

// Produces the DSN (data source name) for the read replica for opening a postgres db connection.
func getPostgresReadDsn(ctx context.Context, pgConfig PostgresConfig) string {
	return postgresDsn(ctx, pgConfig, pgConfig.ReadReplicaHost)
}

func postgresDsn(ctx context.Context, pgConfig PostgresConfig, host string) string {
	if pgConfig.IsIAMAuth() {
		// Checked before the password lookup so IAM never takes the password-less branch below,
		// which forces sslmode=disable and drops ExtraOptions.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s %s",
			host, pgConfig.Port, pgConfig.DbName, pgConfig.User, pgConfig.ExtraOptions)
	}
	password := resolvePassword(ctx, pgConfig.Password, pgConfig.PasswordPath)
	if len(password) == 0 {
		// The password-less case is included for development environments.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			host, pgConfig.Port, pgConfig.DbName, pgConfig.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		host, pgConfig.Port, pgConfig.DbName, pgConfig.User, password, pgConfig.ExtraOptions)
}

// connectorOptions validates pgConfig and returns the pgx connector options it needs. With IAM auth
// that is a BeforeConnect hook that attaches a fresh token to every new physical connection.
func connectorOptions(ctx context.Context, pgConfig PostgresConfig) ([]stdlib.OptionOpenDB, error) {
	if err := pgConfig.Validate(); err != nil {
		return nil, err
	}
	if !pgConfig.IsIAMAuth() {
		return nil, nil
	}
	buildToken, err := newIAMTokenBuilder(ctx, pgConfig.Region)
	if err != nil {
		return nil, err
	}
	return []stdlib.OptionOpenDB{stdlib.OptionBeforeConnect(iamBeforeConnect(buildToken))}, nil
}

// CreatePostgresDbIfNotExists creates DB if it doesn't exist for the passed in config
func CreatePostgresDbIfNotExists(ctx context.Context, pgConfig PostgresConfig, options ...otelsql.Option) (*sqlx.DB, error) {
	connOpts, err := connectorOptions(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	dsn := GetPostgresDsn(ctx, pgConfig)
	db, err := openPostgres(dsn, connOpts, options...)
	if err != nil {
		return nil, err
	}

	// sqlx.Open doesn't actually connect; we need to ping to verify the connection.
	if err = db.PingContext(ctx); err == nil {
		return db, nil
	}

	if !IsPgErrorWithCode(err, PqInvalidDBCode) {
		logger.Errorf(ctx, "Unhandled error connecting to postgres, pg [%v]: %v", pgConfig, err)
		db.Close() //nolint:errcheck
		return nil, err
	}
	db.Close() //nolint:errcheck
	logger.Warningf(ctx, "Database [%v] does not exist", pgConfig.DbName)

	// Every postgres installation includes a 'postgres' database by default. We connect to that now in order to
	// initialize the user-specified database.
	defaultDbPgConfig := pgConfig
	defaultDbPgConfig.DbName = defaultDB
	defaultDsn := GetPostgresDsn(ctx, defaultDbPgConfig)
	defaultDb, err := openPostgres(defaultDsn, connOpts, options...)
	if err != nil {
		return nil, err
	}
	defer defaultDb.Close() //nolint: errcheck

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
	db, err = openPostgres(dsn, connOpts, options...)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		db.Close() //nolint:errcheck
		return nil, err
	}

	return db, nil
}

// CreatePostgresReadOnlyDbConnection creates readonly DB connection and returns the sqlx.DB object and error
func CreatePostgresReadOnlyDbConnection(ctx context.Context, pgConfig PostgresConfig, options ...otelsql.Option) (*sqlx.DB, error) {
	connOpts, err := connectorOptions(ctx, pgConfig)
	if err != nil {
		return nil, err
	}

	dsn := getPostgresReadDsn(ctx, pgConfig)
	db, err := openPostgres(dsn, connOpts, options...)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		db.Close() //nolint: errcheck
		return nil, err
	}

	return db, nil
}

// NewPostgresListener returns a lib/pq LISTEN/NOTIFY listener for pgConfig's primary host.
//
// pq.Listener redials with a fixed DSN and has no per-connection hook, so with IAM auth the token is kept in
// a private pgpass file instead, which lib/pq reads on every dial. The file is rewritten before the first
// connection and on every disconnect or failed attempt, which pq.Listener reports synchronously just before
// it redials. The file stays in a private temp directory for the life of the process.
func NewPostgresListener(ctx context.Context, pgConfig PostgresConfig, minReconnect, maxReconnect time.Duration,
	callback pq.EventCallbackType) (*pq.Listener, error) {

	if err := pgConfig.Validate(); err != nil {
		return nil, err
	}
	dsn := GetPostgresDsn(ctx, pgConfig)
	if !pgConfig.IsIAMAuth() {
		return pq.NewListener(dsn, minReconnect, maxReconnect, callback), nil
	}

	buildToken, err := newIAMTokenBuilder(ctx, pgConfig.Region)
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "flyte-pgpass-")
	if err != nil {
		return nil, fmt.Errorf("create pgpass directory for RDS IAM auth: %w", err)
	}
	passfile := &iamPassfile{
		path:       filepath.Join(dir, "pgpass"),
		endpoint:   net.JoinHostPort(pgConfig.Host, strconv.Itoa(pgConfig.Port)),
		user:       pgConfig.User,
		buildToken: buildToken,
	}
	if err := passfile.refresh(ctx); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
	}

	dsn += " passfile=" + quoteDsnValue(passfile.path)
	return pq.NewListener(dsn, minReconnect, maxReconnect, func(ev pq.ListenerEventType, err error) {
		if ev == pq.ListenerEventDisconnected || ev == pq.ListenerEventConnectionAttemptFailed {
			if refreshErr := passfile.refresh(context.Background()); refreshErr != nil {
				logger.Errorf(context.Background(), "Failed to refresh the listener's RDS IAM auth token: %v", refreshErr)
			}
		}
		if callback != nil {
			callback(ev, err)
		}
	}), nil
}

func IsPgErrorWithCode(err error, code string) bool {
	pgErr := &pgconn.PgError{}
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}

	return false
}
