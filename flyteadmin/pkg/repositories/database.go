package repositories

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// getGormLogLevel converts between the flytestdlib configured log level to the equivalent gorm log level.
func getGormLogLevel(ctx context.Context, logConfig *logger.Config) gormLogger.LogLevel {
	if logConfig == nil {
		logger.Debugf(ctx, "No log config block found, setting gorm db log level to: error")
		return gormLogger.Error
	}
	switch logConfig.Level {
	case logger.PanicLevel:
		fallthrough
	case logger.FatalLevel:
		fallthrough
	case logger.ErrorLevel:
		return gormLogger.Error
	case logger.WarnLevel:
		return gormLogger.Warn
	case logger.InfoLevel:
		fallthrough
	case logger.DebugLevel:
		return gormLogger.Info
	default:
		return gormLogger.Silent
	}
}

// Resolves a password value from either a user-provided inline value or a filepath whose contents contain a password.
func resolvePassword(ctx context.Context, passwordVal, passwordPath string) string {
	password := passwordVal
	if len(passwordPath) > 0 {
		if _, err := os.Stat(passwordPath); os.IsNotExist(err) {
			logger.Fatalf(ctx,
				"missing database password at specified path [%s]", passwordPath)
		}
		passwordVal, err := ioutil.ReadFile(passwordPath)
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
func getPostgresDsn(ctx context.Context, pgConfig runtimeInterfaces.PostgresConfig) string {
	password := resolvePassword(ctx, pgConfig.Password, pgConfig.PasswordPath)
	if len(password) == 0 {
		// The password-less case is included for development environments.
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		pgConfig.Host, pgConfig.Port, pgConfig.DbName, pgConfig.User, password, pgConfig.ExtraOptions)
}

func GetDB(ctx context.Context, dbConfig *runtimeInterfaces.DbConfig, logConfig *logger.Config) (
	*gorm.DB, error) {
	if dbConfig == nil {
		panic("Cannot initialize database repository from empty db config")
	}
	var dialector gorm.Dialector
	logLevel := getGormLogLevel(ctx, logConfig)

	switch {
	// TODO: Figure out a better proxy for a non-empty postgres config
	case len(dbConfig.PostgresConfig.Host) > 0 || len(dbConfig.PostgresConfig.User) > 0 || len(dbConfig.PostgresConfig.DbName) > 0:
		dialector = postgres.Open(getPostgresDsn(ctx, dbConfig.PostgresConfig))
		// TODO: add other gorm-supported db type handling in further case blocks.
	case len(dbConfig.DeprecatedHost) > 0 || len(dbConfig.DeprecatedUser) > 0 || len(dbConfig.DeprecatedDbName) > 0:
		pgConfig := runtimeInterfaces.PostgresConfig{
			Host:         dbConfig.DeprecatedHost,
			Port:         dbConfig.DeprecatedPort,
			DbName:       dbConfig.DeprecatedDbName,
			User:         dbConfig.DeprecatedUser,
			Password:     dbConfig.DeprecatedPassword,
			PasswordPath: dbConfig.DeprecatedPasswordPath,
			ExtraOptions: dbConfig.DeprecatedExtraOptions,
			Debug:        dbConfig.DeprecatedDebug,
		}
		dialector = postgres.Open(getPostgresDsn(ctx, pgConfig))
	default:
		panic(fmt.Sprintf("Unrecognized database config %v", dbConfig))
	}

	return gorm.Open(dialector, &gorm.Config{
		Logger:                                   gormLogger.Default.LogMode(logLevel),
		DisableForeignKeyConstraintWhenMigrating: !dbConfig.EnableForeignKeyConstraintWhenMigrating,
	})
}
