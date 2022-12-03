package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/stretchr/testify/assert"
)

func TestParseDatabaseConfig(t *testing.T) {
	assert.NoError(t, logger.SetConfig(&logger.Config{IncludeSourceCode: true}))

	accessor := viper.NewAccessor(config.Options{
		RootSection: configSection,
		SearchPaths: []string{filepath.Join("testdata", "config.yaml")},
	})

	assert.NoError(t, accessor.UpdateConfig(context.Background()))

	assert.Equal(t, false, GetConfig().EnableForeignKeyConstraintWhenMigrating)
	assert.Equal(t, 100, GetConfig().MaxOpenConnections)
	assert.Equal(t, 10, GetConfig().MaxIdleConnections)
	assert.Equal(t, config.Duration{Duration: 3600000000000}, GetConfig().ConnMaxLifeTime)

	assert.Equal(t, false, GetConfig().Postgres.IsEmpty())
	assert.Equal(t, 5432, GetConfig().Postgres.Port)
	assert.Equal(t, "postgres", GetConfig().Postgres.User)
	assert.Equal(t, "postgres", GetConfig().Postgres.Host)
	assert.Equal(t, "postgres", GetConfig().Postgres.DbName)
	assert.Equal(t, "sslmode=disable", GetConfig().Postgres.ExtraOptions)
	assert.Equal(t, "password", GetConfig().Postgres.Password)
	assert.Equal(t, "/etc/secret", GetConfig().Postgres.PasswordPath)
	assert.Equal(t, true, GetConfig().Postgres.Debug)

	assert.Equal(t, false, GetConfig().SQLite.IsEmpty())
	assert.Equal(t, "admin.db", GetConfig().SQLite.File)
}
