package gormimpl

import (
	"testing"

	"context"

	mocket "github.com/Selvatico/go-mocket"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	datacatalog_error "github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
)

func getTestDataset() models.Dataset {
	return models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: "testProject",
			Domain:  "testDomain",
			Name:    "testName",
			Version: "testVersion",
		},
		SerializedMetadata: []byte{1, 2, 3},
	}
}

func TestCreateDataset(t *testing.T) {
	dataset := getTestDataset()
	datasetCreated := false
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "datasets" ("created_at","updated_at","deleted_at","project","name","domain","version","serialized_metadata") VALUES (?,?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			assert.EqualValues(t, dataset.Project, values[3].Value)
			assert.EqualValues(t, dataset.Name, values[4].Value)
			assert.EqualValues(t, dataset.Domain, values[5].Value)
			assert.EqualValues(t, dataset.Version, values[6].Value)
			assert.EqualValues(t, dataset.SerializedMetadata, values[7].Value)
			datasetCreated = true
		},
	)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	err := datasetRepo.Create(context.Background(), getTestDataset())
	assert.NoError(t, err)
	assert.True(t, datasetCreated)
}

func TestGetDataset(t *testing.T) {
	dataset := getTestDataset()
	expectedResponse := make([]map[string]interface{}, 0)
	sampleDataset := make(map[string]interface{})
	sampleDataset["project"] = dataset.Project
	sampleDataset["domain"] = dataset.Domain
	sampleDataset["name"] = dataset.Name
	sampleDataset["version"] = dataset.Version

	expectedResponse = append(expectedResponse, sampleDataset)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "datasets"  WHERE "datasets"."deleted_at" IS NULL AND (("datasets"."project" = testProject) AND ("datasets"."name" = testName) AND ("datasets"."domain" = testDomain) AND ("datasets"."version" = testVersion)) ORDER BY "datasets"."project" ASC LIMIT 1`).WithReply(expectedResponse)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	actualDataset, err := datasetRepo.Get(context.Background(), dataset.DatasetKey)
	assert.NoError(t, err)
	assert.Equal(t, dataset.Project, actualDataset.Project)
	assert.Equal(t, dataset.Domain, actualDataset.Domain)
	assert.Equal(t, dataset.Name, actualDataset.Name)
	assert.Equal(t, dataset.Version, actualDataset.Version)
}

func TestGetDatasetNotFound(t *testing.T) {
	dataset := getTestDataset()
	sampleDataset := make(map[string]interface{})
	sampleDataset["project"] = dataset.Project
	sampleDataset["domain"] = dataset.Domain
	sampleDataset["name"] = dataset.Name
	sampleDataset["version"] = dataset.Version

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "datasets"  WHERE "datasets"."deleted_at" IS NULL AND (("datasets"."project" = testProject) AND ("datasets"."name" = testName) AND ("datasets"."domain" = testDomain) AND ("datasets"."version" = testVersion)) ORDER BY "datasets"."id" ASC LIMIT 1`).WithReply(nil)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	_, err := datasetRepo.Get(context.Background(), dataset.DatasetKey)
	assert.Error(t, err)
	notFoundErr, ok := err.(datacatalog_error.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, notFoundErr.Code())
}

func TestCreateDatasetAlreadyExists(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "datasets" ("created_at","updated_at","deleted_at","project","name","domain","version","serialized_metadata") VALUES (?,?,?,?,?,?,?,?)`).WithError(
		getAlreadyExistsErr(),
	)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	err := datasetRepo.Create(context.Background(), getTestDataset())
	assert.Error(t, err)
	dcErr, ok := err.(datacatalog_error.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}
