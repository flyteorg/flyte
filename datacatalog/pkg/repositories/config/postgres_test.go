package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/database"
	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func TestSetupDbConnectionPool(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		gormDb, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), &gorm.Config{})
		assert.Nil(t, err)
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
			ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
		}
		err = setupDbConnectionPool(gormDb, dbConfig)
		assert.Nil(t, err)
		genericDb, err := gormDb.DB()
		assert.Nil(t, err)
		assert.Equal(t, genericDb.Stats().MaxOpenConnections, 1000)
	})
	t.Run("unset duration", func(t *testing.T) {
		gormDb, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), &gorm.Config{})
		assert.Nil(t, err)
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
		}
		err = setupDbConnectionPool(gormDb, dbConfig)
		assert.Nil(t, err)
		genericDb, err := gormDb.DB()
		assert.Nil(t, err)
		assert.Equal(t, genericDb.Stats().MaxOpenConnections, 1000)
	})
	t.Run("failed to get DB", func(t *testing.T) {
		gormDb := &gorm.DB{
			Config: &gorm.Config{
				ConnPool: &gorm.PreparedStmtDB{},
			},
		}
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
			ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
		}
		err := setupDbConnectionPool(gormDb, dbConfig)
		assert.NotNil(t, err)
	})
}
