package config

import (
	"testing"

	"gorm.io/gorm/logger"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

func TestConstructGormArgs(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		ExtraOptions: "sslmode=disable",
		BaseConfig: BaseConfig{
			LogLevel:                                 3,
			DisableForeignKeyConstraintWhenMigrating: true,
		},
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", postgresConfigProvider.GetDSN())
	assert.Equal(t, logger.LogLevel(3), postgresConfigProvider.GetDBConfig().LogLevel)
	assert.Equal(t, true, postgresConfigProvider.GetDBConfig().DisableForeignKeyConstraintWhenMigrating)
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
