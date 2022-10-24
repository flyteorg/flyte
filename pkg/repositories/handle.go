package repositories

import (
	"context"

	"gorm.io/driver/sqlite"

	"fmt"

	"github.com/flyteorg/flytestdlib/database"

	"github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"
)

type DBHandle struct {
	db *gorm.DB
}

func NewDBHandle(ctx context.Context, dbConfigValues database.DbConfig, catalogScope promutils.Scope) (*DBHandle, error) {
	var gormDb *gorm.DB
	var err error

	switch {
	case !dbConfigValues.SQLite.IsEmpty():
		gormDb, err = gorm.Open(sqlite.Open(dbConfigValues.SQLite.File))
	case !dbConfigValues.Postgres.IsEmpty():
		gormDb, err = config.OpenDbConnection(ctx, config.NewPostgresConfigProvider(dbConfigValues, catalogScope.NewSubScope(config.Postgres)))
	default:
		return nil, fmt.Errorf("unrecognized database config, %v. Supported only postgres and sqlite", dbConfigValues)
	}

	if err != nil {
		return nil, err
	}

	out := &DBHandle{
		db: gormDb,
	}

	return out, nil
}

func (h *DBHandle) CreateDB(dbName string) error {
	type DatabaseResult struct {
		Exists bool
	}
	var checkExists DatabaseResult
	result := h.db.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", dbName).Scan(&checkExists)
	if result.Error != nil {
		return result.Error
	}

	// create db if it does not exist
	if !checkExists.Exists {
		logger.Infof(context.TODO(), "Creating Database %v since it does not exist", dbName)

		// NOTE: golang sql drivers do not support parameter injection for CREATE calls
		createDBStatement := fmt.Sprintf("CREATE DATABASE %s", dbName)
		result = h.db.Exec(createDBStatement)

		if result.Error != nil {
			if !isPgErrorWithCode(result.Error, pqDbAlreadyExistsCode) {
				return result.Error
			}
			logger.Infof(context.TODO(), "Not creating database %s, already exists", dbName)
		}
	}

	return nil
}

func (h *DBHandle) Migrate(ctx context.Context) error {
	if err := h.db.AutoMigrate(&models.Dataset{}); err != nil {
		return err
	}

	if err := h.db.Debug().AutoMigrate(&models.Artifact{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.ArtifactData{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.Tag{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.PartitionKey{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.Partition{}); err != nil {
		return err
	}

	if err := h.db.AutoMigrate(&models.Reservation{}); err != nil {
		return err
	}

	return nil
}
