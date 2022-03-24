package repositories

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
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
	pgConfig := &runtimeInterfaces.PostgresConfig{
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
	t.Run("successful", func(t *testing.T) {
		gormDb, err := gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.db")), &gorm.Config{})
		assert.Nil(t, err)
		dbConfig := &runtimeInterfaces.DbConfig{
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
		dbConfig := &runtimeInterfaces.DbConfig{
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
		dbConfig := &runtimeInterfaces.DbConfig{
			DeprecatedPort:     5432,
			MaxIdleConnections: 10,
			MaxOpenConnections: 1000,
			ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
		}
		err := setupDbConnectionPool(gormDb, dbConfig)
		assert.NotNil(t, err)
	})
}

func TestGetDB(t *testing.T) {
	ctx := context.TODO()

	t.Run("missing DB Config", func(t *testing.T) {
		_, err := GetDB(ctx, &runtimeInterfaces.DbConfig{}, &logger.Config{})
		assert.Error(t, err)
	})

	t.Run("sqlite config", func(t *testing.T) {
		dbFile := path.Join(t.TempDir(), "admin.db")
		db, err := GetDB(ctx, &runtimeInterfaces.DbConfig{
			SQLiteConfig: &runtimeInterfaces.SQLiteConfig{
				File: dbFile,
			},
		}, &logger.Config{})
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.FileExists(t, dbFile)
		assert.Equal(t, "sqlite", db.Name())
	})
}
