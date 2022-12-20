package config

import (
	"testing"

	"github.com/flyteorg/flytestdlib/database"
	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

func TestConstructGormArgs(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(database.DbConfig{Postgres: database.PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		ExtraOptions: "sslmode=disable",
	},
		EnableForeignKeyConstraintWhenMigrating: false,
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", postgresConfigProvider.GetDSN())
	assert.Equal(t, false, postgresConfigProvider.GetDBConfig().EnableForeignKeyConstraintWhenMigrating)
}

func TestConstructGormArgsWithPassword(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(database.DbConfig{Postgres: database.PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		Password:     "pass",
		ExtraOptions: "sslmode=enable",
	},
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass sslmode=enable", postgresConfigProvider.GetDSN())
}

func TestConstructGormArgsWithPasswordNoExtra(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(database.DbConfig{Postgres: database.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		DbName:   "postgres",
		User:     "postgres",
		Password: "pass",
	},
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass ", postgresConfigProvider.GetDSN())
}
