package config

import (
	"testing"

	mockScope "github.com/lyft/flytestdlib/promutils"

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
