package gormimpl

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestGetActive(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	createdAt := time.Date(2021, time.June, 1, 00, 00, 00, 00, time.UTC)
	updatedAt := time.Date(2022, time.June, 1, 00, 00, 00, 00, time.UTC)
	expectedMetadata := []map[string]interface{}{{
		"id":                1,
		"created_at":        createdAt,
		"updated_at":        updatedAt,
		"deleted_at":        nil,
		"version":           "v1",
		"document_location": "s3://bucket/key",
		"active":            true,
	}}
	queried := false
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "configuration_document_metadata" WHERE "configuration_document_metadata"."active" = $1 LIMIT 1`).WithArgs(true).WithReply(expectedMetadata).WithCallback(
		func(s string, values []driver.NamedValue) {
			queried = true
		},
	)

	metadata, err := configurationRepo.GetActive(context.Background())
	assert.True(t, queried)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), metadata.ID)
	assert.Equal(t, createdAt.String(), metadata.CreatedAt.String())
	assert.Equal(t, updatedAt.String(), metadata.UpdatedAt.String())
	assert.Nil(t, metadata.DeletedAt)
	assert.Equal(t, "v1", metadata.Version)
	assert.Equal(t, storage.DataReference("s3://bucket/key"), metadata.DocumentLocation)
	assert.True(t, metadata.Active)
}

func TestGetActiveNotFound(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "configuration_document_metadata" WHERE "configuration_document_metadata"."active" = $1 LIMIT 1`).WithArgs(true).WithError(gorm.ErrRecordNotFound)

	metadata, err := configurationRepo.GetActive(context.Background())
	assert.Equal(t, models.ConfigurationDocumentMetadata{}, metadata)
	assert.EqualError(t, err, "missing singleton entity of type configuration")
}

func TestGetActiveDBError(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "configuration_document_metadata" WHERE "configuration_document_metadata"."active" = $1 LIMIT 1`).WithArgs(true).WithError(gorm.ErrUnsupportedDriver)

	metadata, err := configurationRepo.GetActive(context.Background())
	assert.Equal(t, models.ConfigurationDocumentMetadata{}, metadata)
	assert.EqualError(t, err, "Test transformer failed to find transformation to apply")
}

func TestCreate(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "configuration_document_metadata" WHERE "configuration_document_metadata"."active" = $1 LIMIT 1`).WithArgs(true).WithError(gorm.ErrRecordNotFound)
	created := false
	GlobalMock.NewMock().WithQuery(`INSERT INTO "configuration_document_metadata" ("created_at","updated_at","deleted_at","version","document_location","active") VALUES ($1,$2,$3,$4,$5,$6)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			created = true
		},
	)
	err := configurationRepo.Create(context.Background(), &models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: "s3://bucket/key",
		Active:           true,
	})
	assert.NoError(t, err)
	assert.True(t, created)
}

func TestCreateActiveDocumentExistsSameVersion(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "configuration_document_metadata" WHERE "configuration_document_metadata"."active" = $1 LIMIT 1`).WithArgs(true).WithReply([]map[string]interface{}{{"version": "v1"}})
	created := false
	GlobalMock.NewMock().WithQuery(`INSERT INTO "configuration_document_metadata" ("created_at","updated_at","deleted_at","version","document_location","active") VALUES ($1,$2,$3,$4,$5,$6)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			created = true
		},
	)

	err := configurationRepo.Create(context.Background(), &models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: "s3://bucket/key",
		Active:           true,
	})
	assert.False(t, created)
	assert.Nil(t, err)
}

func TestCreateActiveDocumentExistsDifferentVersion(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "configuration_document_metadata" WHERE "configuration_document_metadata"."active" = $1 LIMIT 1`).WithArgs(true).WithReply([]map[string]interface{}{{"version": "v2"}})

	err := configurationRepo.Create(context.Background(), &models.ConfigurationDocumentMetadata{
		Version:          "v1",
		DocumentLocation: "s3://bucket/key",
		Active:           true,
	})
	assert.EqualError(t, err, "There is already an active configuration document.")
}

func TestUpdate(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	updated := false
	GlobalMock.NewMock().WithQuery(`UPDATE "configuration_document_metadata" SET "active"=$1,"updated_at"=$2 WHERE "configuration_document_metadata"."version" = $3 AND "configuration_document_metadata"."active" = $4`).WithRowsNum(1).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)

	created := false
	GlobalMock.NewMock().WithQuery(`INSERT INTO "configuration_document_metadata" ("created_at","updated_at","deleted_at","version","document_location","active") VALUES ($1,$2,$3,$4,$5,$6)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			created = true
		},
	)

	err := configurationRepo.Update(context.Background(), &interfaces.UpdateConfigurationInput{
		VersionToUpdate: "v1",
		NewConfigurationMetadata: &models.ConfigurationDocumentMetadata{
			Version:          "v2",
			DocumentLocation: "s3://bucket/key",
			Active:           true,
		},
	})
	assert.True(t, updated)
	assert.True(t, created)
	assert.NoError(t, err)
}

func TestUpdateStale(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	updated := false
	GlobalMock.NewMock().WithQuery(`UPDATE "configuration_document_metadata" SET "active"=$1,"updated_at"=$2 WHERE "configuration_document_metadata"."version" = $3 AND "configuration_document_metadata"."active" = $4`).WithRowsNum(0).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)

	err := configurationRepo.Update(context.Background(), &interfaces.UpdateConfigurationInput{
		VersionToUpdate: "v1",
		NewConfigurationMetadata: &models.ConfigurationDocumentMetadata{
			Version:          "v2",
			DocumentLocation: "s3://bucket/key",
			Active:           true,
		},
	})
	assert.True(t, updated)
	assert.EqualError(t, err, "The document you are trying to update is outdated. Please try again.")
}

func TestUpdateDBError(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(`UPDATE "configuration_document_metadata" SET "active"=$1,"updated_at"=$2 WHERE "configuration_document_metadata"."version" = $3 AND "configuration_document_metadata"."active" = $4`).WithError(gorm.ErrUnsupportedDriver)

	err := configurationRepo.Update(context.Background(), &interfaces.UpdateConfigurationInput{
		VersionToUpdate: "v1",
		NewConfigurationMetadata: &models.ConfigurationDocumentMetadata{
			Version:          "v2",
			DocumentLocation: "s3://bucket/key",
			Active:           true,
		},
	})
	assert.EqualError(t, err, "Test transformer failed to find transformation to apply")
}

func TestUpdateDBError2(t *testing.T) {
	configurationRepo := NewConfigurationRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	updated := false
	GlobalMock.NewMock().WithQuery(`UPDATE "configuration_document_metadata" SET "active"=$1,"updated_at"=$2 WHERE "configuration_document_metadata"."version" = $3 AND "configuration_document_metadata"."active" = $4`).WithRowsNum(1).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)
	GlobalMock.NewMock().WithQuery(`INSERT INTO "configuration_document_metadata" ("created_at","updated_at","deleted_at","version","document_location","active") VALUES ($1,$2,$3,$4,$5,$6)`).WithError(gorm.ErrUnsupportedDriver)

	err := configurationRepo.Update(context.Background(), &interfaces.UpdateConfigurationInput{
		VersionToUpdate: "v1",
		NewConfigurationMetadata: &models.ConfigurationDocumentMetadata{
			Version:          "v2",
			DocumentLocation: "s3://bucket/key",
			Active:           true,
		},
	})
	assert.True(t, updated)
	assert.EqualError(t, err, "Test transformer failed to find transformation to apply")
}
