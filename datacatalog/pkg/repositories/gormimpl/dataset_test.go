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
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func getTestDataset() models.Dataset {
	return models.Dataset{
		DatasetKey: models.DatasetKey{
			Project: "testProject",
			Domain:  "testDomain",
			Name:    "testName",
			Version: "testVersion",
		},
		SerializedMetadata: []byte{1, 2, 3},
		PartitionKeys: []models.PartitionKey{
			{Name: "key1"},
			{Name: "key2"},
		},
	}
}

// sql will generate a uuid
func getDatasetUUID() string {
	return "test-uuid"
}

func TestCreateDatasetNoPartitions(t *testing.T) {
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

	dataset.PartitionKeys = nil

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := datasetRepo.Create(context.Background(), dataset)
	assert.NoError(t, err)
	assert.True(t, datasetCreated)
}

func TestCreateDataset(t *testing.T) {
	dataset := getTestDataset()
	datasetCreated := false
	insertKeyQueryNum := 0
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

	GlobalMock.NewMock().WithQuery(
		`SELECT "uuid" FROM "datasets"  WHERE (project = testProject) AND (name = testName) AND (domain = testDomain) AND (version = testVersion)`).WithReply([]map[string]interface{}{{"uuid": getDatasetUUID()}})

	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "partition_keys" ("created_at","updated_at","deleted_at","dataset_uuid","name") VALUES (?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			assert.EqualValues(t, getDatasetUUID(), values[3].Value)
			assert.EqualValues(t, dataset.PartitionKeys[insertKeyQueryNum].Name, values[4].Value)
			insertKeyQueryNum++
		},
	)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := datasetRepo.Create(context.Background(), getTestDataset())
	assert.NoError(t, err)
	assert.True(t, datasetCreated)
	assert.Equal(t, insertKeyQueryNum, 2)
}

func TestGetDataset(t *testing.T) {
	dataset := getTestDataset()

	expectedDatasetResponse := make([]map[string]interface{}, 0)
	sampleDataset := make(map[string]interface{})
	sampleDataset["project"] = dataset.Project
	sampleDataset["domain"] = dataset.Domain
	sampleDataset["name"] = dataset.Name
	sampleDataset["version"] = dataset.Version
	sampleDataset["uuid"] = getDatasetUUID()

	expectedDatasetResponse = append(expectedDatasetResponse, sampleDataset)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "datasets"  WHERE "datasets"."deleted_at" IS NULL AND (("datasets"."project" = testProject) AND ("datasets"."name" = testName) AND ("datasets"."domain" = testDomain) AND ("datasets"."version" = testVersion)) ORDER BY "datasets"."project" ASC LIMIT 1`).WithReply(expectedDatasetResponse)

	expectedPartitionKeyResponse := make([]map[string]interface{}, 0)
	samplePartitionKey := make(map[string]interface{})
	samplePartitionKey["name"] = "testKey1"
	samplePartitionKey["dataset_uuid"] = getDatasetUUID()
	expectedPartitionKeyResponse = append(expectedPartitionKeyResponse, samplePartitionKey, samplePartitionKey)

	GlobalMock.NewMock().WithQuery(`SELECT * FROM "partition_keys"  WHERE "partition_keys"."deleted_at" IS NULL AND (("dataset_uuid" IN (test-uuid))) ORDER BY "partition_keys"."dataset_uuid" ASC`).WithReply(expectedPartitionKeyResponse)
	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	actualDataset, err := datasetRepo.Get(context.Background(), dataset.DatasetKey)
	assert.NoError(t, err)
	assert.Equal(t, dataset.Project, actualDataset.Project)
	assert.Equal(t, dataset.Domain, actualDataset.Domain)
	assert.Equal(t, dataset.Name, actualDataset.Name)
	assert.Equal(t, dataset.Version, actualDataset.Version)
	assert.Equal(t, getDatasetUUID(), actualDataset.UUID)
	assert.Len(t, actualDataset.PartitionKeys, 2)
}

func TestGetDatasetWithUUID(t *testing.T) {
	dataset := models.Dataset{
		DatasetKey: models.DatasetKey{
			UUID: getDatasetUUID(),
		},
	}
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
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "datasets"  WHERE "datasets"."deleted_at" IS NULL AND (("datasets"."uuid" = test-uuid)) ORDER BY "datasets"."project" ASC LIMIT 1`).WithReply(expectedResponse)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
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

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
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

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := datasetRepo.Create(context.Background(), getTestDataset())
	assert.Error(t, err)
	dcErr, ok := err.(datacatalog_error.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}
