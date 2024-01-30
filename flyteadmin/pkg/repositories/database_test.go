package repositories

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func TestResolvePassword(t *testing.T) {
	password := "123abc"
	tmpFile, err := ioutil.TempFile("", "prefix")
	if err != nil {
		t.Errorf("Couldn't open temp file: %v", err)
	}
	defer tmpFile.Close()
	if _, err = tmpFile.WriteString(password); err != nil {
		t.Errorf("Couldn't write to temp file: %v", err)
	}
	resolvedPassword := resolvePassword(context.TODO(), "", tmpFile.Name())
	assert.Equal(t, resolvedPassword, password)
}

func TestGetPostgresDsn(t *testing.T) {
	pgConfig := database.PostgresConfig{
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
	t.Run("with password path", func(t *testing.T) {
		password := "123abc"
		tmpFile, err := ioutil.TempFile("", "prefix")
		if err != nil {
			t.Errorf("Couldn't open temp file: %v", err)
		}
		defer tmpFile.Close()
		if _, err = tmpFile.WriteString(password); err != nil {
			t.Errorf("Couldn't write to temp file: %v", err)
		}
		pgConfig.PasswordPath = tmpFile.Name()
		dsn := getPostgresDsn(context.TODO(), pgConfig)
		assert.Equal(t, "host=localhost port=5432 dbname=postgres user=postgres password=123abc ", dsn)
	})
}

func TestSetupDbConnectionPool(t *testing.T) {
	ctx := context.TODO()
	t.Run("successful", func(t *testing.T) {
		gormDb, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), &gorm.Config{})
		assert.Nil(t, err)
		dbConfig := &database.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
			ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
		}
		err = setupDbConnectionPool(ctx, gormDb, dbConfig)
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
		err = setupDbConnectionPool(ctx, gormDb, dbConfig)
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
		err := setupDbConnectionPool(ctx, gormDb, dbConfig)
		assert.NotNil(t, err)
	})
}

func TestGetDB(t *testing.T) {
	ctx := context.TODO()

	t.Run("missing DB Config", func(t *testing.T) {
		_, err := GetDB(ctx, &database.DbConfig{}, &logger.Config{})
		assert.Error(t, err)
	})

	t.Run("sqlite config", func(t *testing.T) {
		dbFile := path.Join(t.TempDir(), "admin.db")
		db, err := GetDB(ctx, &database.DbConfig{
			SQLite: database.SQLiteConfig{
				File: dbFile,
			},
		}, &logger.Config{})
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.FileExists(t, dbFile)
		assert.Equal(t, "sqlite", db.Name())
	})
}
