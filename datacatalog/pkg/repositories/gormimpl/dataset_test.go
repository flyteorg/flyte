package gormimpl

import (
	"context"
	"testing"
	"time"

	mocket "github.com/Selvatico/go-mocket"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	"github.com/flyteorg/datacatalog/pkg/common"
	datacatalog_error "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/utils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
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
		BaseModel: models.BaseModel{
			CreatedAt: time.Unix(111, 0),
			UpdatedAt: time.Unix(111, 0),
			DeletedAt: nil,
		},
	}
}

// Raw db response to return on raw queries for datasets
func getDBDatasetResponse(dataset models.Dataset) []map[string]interface{} {
	expectedDatasetResponse := make([]map[string]interface{}, 0)
	sampleDataset := make(map[string]interface{})
	sampleDataset["project"] = dataset.Project
	sampleDataset["domain"] = dataset.Domain
	sampleDataset["name"] = dataset.Name
	sampleDataset["version"] = dataset.Version
	sampleDataset["uuid"] = getDatasetUUID()
	expectedDatasetResponse = append(expectedDatasetResponse, sampleDataset)
	return expectedDatasetResponse
}

func getDBPartitionKeysResponse(datasets []models.Dataset) []map[string]interface{} {
	expectedPartitionKeyResponse := make([]map[string]interface{}, 0)

	for _, dataset := range datasets {
		samplePartitionKey := make(map[string]interface{})
		samplePartitionKey["name"] = "key1"
		samplePartitionKey["dataset_uuid"] = dataset.UUID
		expectedPartitionKeyResponse = append(expectedPartitionKeyResponse, samplePartitionKey)
	}

	return expectedPartitionKeyResponse
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
		`INSERT INTO "datasets" ("created_at","updated_at","deleted_at","project","name","domain","version","serialized_metadata") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`).WithCallback(
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
		`INSERT INTO "datasets" ("created_at","updated_at","deleted_at","project","name","domain","version","serialized_metadata") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			assert.EqualValues(t, dataset.Project, values[3].Value)
			assert.EqualValues(t, dataset.Name, values[4].Value)
			assert.EqualValues(t, dataset.Domain, values[5].Value)
			assert.EqualValues(t, dataset.Version, values[6].Value)
			assert.EqualValues(t, dataset.SerializedMetadata, values[7].Value)
			datasetCreated = true
		},
	).WithReply([]map[string]interface{}{{"dataset_uuid": getDatasetUUID()}})

	GlobalMock.NewMock().WithQuery(
		`INSERT INTO "partition_keys" ("created_at","updated_at","deleted_at","dataset_uuid","name") VALUES ($1,$2,$3,$4,$5),($6,$7,$8,$9,$10) ON CONFLICT ("dataset_uuid","name") DO UPDATE SET "dataset_uuid"="excluded"."dataset_uuid"`).WithCallback(
		func(s string, values []driver.NamedValue) {
			assert.EqualValues(t, getDatasetUUID(), values[3].Value)
			assert.EqualValues(t, dataset.PartitionKeys[insertKeyQueryNum].Name, values[4].Value)
			insertKeyQueryNum += 2 // batch insertion
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
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "datasets" WHERE "datasets"."project" = $1 AND "datasets"."name" = $2 AND "datasets"."domain" = $3 AND "datasets"."version" = $4 ORDER BY "datasets"."created_at" LIMIT 1`).WithReply(expectedDatasetResponse)

	expectedPartitionKeyResponse := make([]map[string]interface{}, 0)
	samplePartitionKey := make(map[string]interface{})
	samplePartitionKey["name"] = "testKey1"
	samplePartitionKey["dataset_uuid"] = getDatasetUUID()
	expectedPartitionKeyResponse = append(expectedPartitionKeyResponse, samplePartitionKey, samplePartitionKey)

	GlobalMock.NewMock().WithQuery(`SELECT * FROM "partition_keys" WHERE "partition_keys"."dataset_uuid" = $1 ORDER BY partition_keys.created_at ASC`).WithReply(expectedPartitionKeyResponse)
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
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "datasets" WHERE "datasets"."uuid" = $1 ORDER BY "datasets"."created_at" LIMIT 1%!(EXTRA string=test-uuid)`).WithReply(expectedResponse)

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
		`INSERT INTO "datasets" ("created_at","updated_at","deleted_at","project","name","domain","version","serialized_metadata") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`).WithError(
		getAlreadyExistsErr(),
	)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := datasetRepo.Create(context.Background(), getTestDataset())
	assert.Error(t, err)
	dcErr, ok := err.(datacatalog_error.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code().String(), codes.AlreadyExists.String())
}

func TestListDatasets(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	dataset := getTestDataset()
	dataset.UUID = getDatasetUUID()
	expectedDatasetDBResponse := getDBDatasetResponse(dataset)

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "datasets" LIMIT 10`).WithReply(expectedDatasetDBResponse)

	expectedPartitionKeyResponse := getDBPartitionKeysResponse([]models.Dataset{dataset})
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "partition_keys" WHERE "partition_keys"."dataset_uuid" = $1`).WithReply(expectedPartitionKeyResponse)
	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	listInput := models.ListModelsInput{
		Limit: 10,
	}
	datasets, err := datasetRepo.List(context.Background(), listInput)
	assert.NoError(t, err)
	assert.Len(t, datasets, 1)
	assert.Equal(t, datasets[0].Project, dataset.Project)
	assert.Equal(t, datasets[0].Domain, dataset.Domain)
	assert.Equal(t, datasets[0].Name, dataset.Name)
	assert.Equal(t, datasets[0].Version, dataset.Version)
	assert.Len(t, datasets[0].PartitionKeys, 1)
	assert.Equal(t, datasets[0].PartitionKeys[0].Name, "key1")
}

func TestListDatasetWithFilter(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	dataset := getTestDataset()
	dataset.UUID = getDatasetUUID()
	expectedDatasetDBResponse := getDBDatasetResponse(dataset)

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "datasets" WHERE datasets.project = $1 AND datasets.domain = $2 ORDER BY datasets.created_at desc LIMIT 10 OFFSET 10`).WithReply(expectedDatasetDBResponse)

	expectedPartitionKeyResponse := getDBPartitionKeysResponse([]models.Dataset{dataset})
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "partition_keys" WHERE "partition_keys"."dataset_uuid" = $1%!(EXTRA string=test-uuid)`).WithReply(expectedPartitionKeyResponse)

	datasetRepo := NewDatasetRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	listInput := models.ListModelsInput{
		ModelFilters: []models.ModelFilter{
			{
				Entity: common.Dataset,
				ValueFilters: []models.ModelValueFilter{
					NewGormValueFilter(common.Equal, "project", "p"),
				},
			},
			{
				Entity: common.Dataset,
				ValueFilters: []models.ModelValueFilter{
					NewGormValueFilter(common.Equal, "domain", "d"),
				},
			},
		},
		Offset:        10,
		Limit:         10,
		SortParameter: NewGormSortParameter(datacatalog.PaginationOptions_CREATION_TIME, datacatalog.PaginationOptions_DESCENDING),
	}
	datasets, err := datasetRepo.List(context.Background(), listInput)
	assert.NoError(t, err)
	assert.Equal(t, datasets[0].Project, dataset.Project)
	assert.Equal(t, datasets[0].Domain, dataset.Domain)
	assert.Equal(t, datasets[0].Name, dataset.Name)
	assert.Equal(t, datasets[0].Version, dataset.Version)
	assert.Len(t, datasets[0].PartitionKeys, 1)
	assert.Equal(t, datasets[0].PartitionKeys[0].Name, "key1")
}
