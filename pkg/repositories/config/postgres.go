package config

import (
	"fmt"

	"gorm.io/gorm/logger"

	"github.com/flyteorg/flytestdlib/promutils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const Postgres = "postgres"

// Generic interface for providing a config necessary to open a database connection.
type DbConnectionConfigProvider interface {
	// Returns database dialector
	GetDialector() gorm.Dialector

	GetDBConfig() DbConfig

	GetDSN() string
}

type BaseConfig struct {
	LogLevel                                 logger.LogLevel `json:"log_level"`
	DisableForeignKeyConstraintWhenMigrating bool
}

// PostgreSQL implementation for DbConnectionConfigProvider.
type PostgresConfigProvider struct {
	config DbConfig
	scope  promutils.Scope
}

// TODO : Make the Config provider itself env based
func NewPostgresConfigProvider(config DbConfig, scope promutils.Scope) DbConnectionConfigProvider {
	return &PostgresConfigProvider{
		config: config,
		scope:  scope,
	}
}

func (p *PostgresConfigProvider) GetDSN() string {
	if p.config.Password == "" {
		// Switch for development
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable",
			p.config.Host, p.config.Port, p.config.DbName, p.config.User)
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
		p.config.Host, p.config.Port, p.config.DbName, p.config.User, p.config.Password, p.config.ExtraOptions)
}

func (p *PostgresConfigProvider) GetDialector() gorm.Dialector {
	return postgres.Open(p.GetDSN())
}

func (p *PostgresConfigProvider) GetDBConfig() DbConfig {
	return p.config
}

// Opens a connection to the database specified in the config.
// You must call CloseDbConnection at the end of your session!
func OpenDbConnection(config DbConnectionConfigProvider) (*gorm.DB, error) {
	db, err := gorm.Open(config.GetDialector(), &gorm.Config{
		Logger:                                   logger.Default.LogMode(config.GetDBConfig().LogLevel),
		DisableForeignKeyConstraintWhenMigrating: config.GetDBConfig().DisableForeignKeyConstraintWhenMigrating,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
