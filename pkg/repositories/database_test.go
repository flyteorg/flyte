package repositories

import (
	"context"
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	gormLogger "gorm.io/gorm/logger"

	"github.com/stretchr/testify/assert"
)

func TestGetGormLogLevel(t *testing.T) {
	assert.Equal(t, gormLogger.Error, getGormLogLevel(context.TODO(), &logger.Config{
		Level: logger.PanicLevel,
	}))
	assert.Equal(t, gormLogger.Error, getGormLogLevel(context.TODO(), &logger.Config{
		Level: logger.FatalLevel,
	}))
	assert.Equal(t, gormLogger.Error, getGormLogLevel(context.TODO(), &logger.Config{
		Level: logger.ErrorLevel,
	}))

	assert.Equal(t, gormLogger.Warn, getGormLogLevel(context.TODO(), &logger.Config{
		Level: logger.WarnLevel,
	}))

	assert.Equal(t, gormLogger.Info, getGormLogLevel(context.TODO(), &logger.Config{
		Level: logger.InfoLevel,
	}))
	assert.Equal(t, gormLogger.Info, getGormLogLevel(context.TODO(), &logger.Config{
		Level: logger.DebugLevel,
	}))

	assert.Equal(t, gormLogger.Error, getGormLogLevel(context.TODO(), nil))
}

func TestGetPostgresDsn(t *testing.T) {
	pgConfig := runtimeInterfaces.PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		DbName:       "postgres",
		User:         "postgres",
		ExtraOptions: "sslmode=disable",
	}
	t.Run("no password", func(t *testing.T) {
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres sslmode=disable", dsn)
	})
	t.Run("with password", func(t *testing.T) {
		pgConfig.Password = "pass"
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass sslmode=disable", dsn)

	})
	t.Run("with password, no extra", func(t *testing.T) {
		pgConfig.Password = "pass"
		pgConfig.ExtraOptions = ""
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=pass ", dsn)
	})
}
