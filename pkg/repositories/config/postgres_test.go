package config

import (
	"testing"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/logger"
)

func TestConstructGormArgs(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		BaseConfig: BaseConfig{
			LogLevel:                                 logger.Info,
			DisableForeignKeyConstraintWhenMigrating: true,
		},
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		ExtraOptions: "sslmode=disable",
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", postgresConfigProvider.GetDSN())
	assert.Equal(t, postgresConfigProvider.GetDBConfig().LogLevel, logger.Info)
}

func TestConstructGormArgsWithPassword(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		Password:     "pass",
		ExtraOptions: "sslmode=enable",
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass sslmode=enable", postgresConfigProvider.GetDSN())
}

func TestConstructGormArgsWithPasswordNoExtra(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		Host:     "localhost",
		Port:     5432,
		DbName:   "postgres",
		User:     "postgres",
		Password: "pass",
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass ", postgresConfigProvider.GetDSN())
}
