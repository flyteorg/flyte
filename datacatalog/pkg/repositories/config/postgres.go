package config

import (
	"context"
	"fmt"

	stdlibLogger "github.com/flyteorg/flytestdlib/logger"

	"gorm.io/gorm/logger"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flytestdlib/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	Postgres = "postgres"
	Sqlite   = "sqlite"
)

// Generic interface for providing a config necessary to open a database connection.
type DbConnectionConfigProvider interface {
	// Returns database dialector
	GetDialector() gorm.Dialector

	GetDBConfig() database.DbConfig

	GetDSN() string
}

type BaseConfig struct {
	LogLevel                                 logger.LogLevel `json:"log_level"`
	DisableForeignKeyConstraintWhenMigrating bool
}

// PostgreSQL implementation for DbConnectionConfigProvider.
type PostgresConfigProvider struct {
	config database.DbConfig
	scope  promutils.Scope
}

// TODO : Make the Config provider itself env based
func NewPostgresConfigProvider(config database.DbConfig, scope promutils.Scope) DbConnectionConfigProvider {
	return &PostgresConfigProvider{
		config: config,
		scope:  scope,
	}
}

func (p *PostgresConfigProvider) GetDSN() string {
	if p.config.Postgres.Password == "" {
		// Switch for development
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			p.config.Postgres.Host, p.config.Postgres.Port, p.config.Postgres.DbName, p.config.Postgres.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		p.config.Postgres.Host, p.config.Postgres.Port, p.config.Postgres.DbName, p.config.Postgres.User, p.config.Postgres.Password, p.config.Postgres.ExtraOptions)
}

func (p *PostgresConfigProvider) GetDialector() gorm.Dialector {
	return postgres.Open(p.GetDSN())
}

func (p *PostgresConfigProvider) GetDBConfig() database.DbConfig {
	return p.config
}

// Opens a connection to the database specified in the config.
// You must call CloseDbConnection at the end of your session!
func OpenDbConnection(ctx context.Context, config DbConnectionConfigProvider) (*gorm.DB, error) {
	db, err := gorm.Open(config.GetDialector(), &gorm.Config{
		Logger:                                   database.GetGormLogger(ctx, stdlibLogger.GetConfig()),
		DisableForeignKeyConstraintWhenMigrating: !config.GetDBConfig().EnableForeignKeyConstraintWhenMigrating,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
