package config

import (
	"testing"

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
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", postgresConfigProvider.GetArgs())
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

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass sslmode=enable", postgresConfigProvider.GetArgs())
}

func TestConstructGormArgsWithPasswordNoExtra(t *testing.T) {
	postgresConfigProvider := NewPostgresConfigProvider(DbConfig{
		Host:     "localhost",
		Port:     5432,
		DbName:   "postgres",
		User:     "postgres",
		Password: "pass",
	}, mockScope.NewTestScope())

	assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass ", postgresConfigProvider.GetArgs())
}
