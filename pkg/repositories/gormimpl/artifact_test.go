package gormimpl

import (
	"testing"

	"context"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	"github.com/lyft/datacatalog/pkg/common"
	apiErrors "github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"google.golang.org/grpc/codes"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func getTestArtifact() models.Artifact {
	return models.Artifact{
		ArtifactKey: models.ArtifactKey{
			ArtifactID:     "123",
			DatasetProject: "testProject",
			DatasetDomain:  "testDomain",
			DatasetName:    "testName",
			DatasetVersion: "testVersion",
		},
		DatasetUUID: "test-uuid",
	}
}

func getTestPartition() models.Partition {
	return models.Partition{
		DatasetUUID: "test-uuid",
		Key:         "region",
		Value:       "value",
		ArtifactID:  "123",
	}
}

// Raw db response to return on raw queries for artifacts
func getDBArtifactResponse(artifact models.Artifact) []map[string]interface{} {
	expectedArtifactResponse := make([]map[string]interface{}, 0)
	sampleArtifact := make(map[string]interface{})
	sampleArtifact["dataset_project"] = artifact.DatasetProject
	sampleArtifact["dataset_domain"] = artifact.DatasetDomain
	sampleArtifact["dataset_name"] = artifact.DatasetName
	sampleArtifact["dataset_version"] = artifact.DatasetVersion
	sampleArtifact["artifact_id"] = artifact.ArtifactID
	sampleArtifact["dataset_uuid"] = artifact.DatasetUUID
	expectedArtifactResponse = append(expectedArtifactResponse, sampleArtifact)
	return expectedArtifactResponse
}

// Raw db response to return on raw queries for the artifact data
func getDBArtifactDataResponse(artifact models.Artifact) []map[string]interface{} {
	expectedArtifactDataResponse := make([]map[string]interface{}, 0)
	sampleArtifactData := make(map[string]interface{})
	sampleArtifactData["dataset_project"] = artifact.DatasetProject
	sampleArtifactData["dataset_domain"] = artifact.DatasetDomain
	sampleArtifactData["dataset_name"] = artifact.DatasetName
	sampleArtifactData["dataset_version"] = artifact.DatasetVersion
	sampleArtifactData["artifact_id"] = artifact.ArtifactID
	sampleArtifactData["name"] = "test-dataloc-name"
	sampleArtifactData["location"] = "test-dataloc-location"
	sampleArtifactData["dataset_uuid"] = artifact.DatasetUUID
	expectedArtifactDataResponse = append(expectedArtifactDataResponse, sampleArtifactData)
	return expectedArtifactDataResponse
}

// Raw db response to return on raw queries for partitions
func getDBPartitionResponse(artifact models.Artifact) []map[string]interface{} {
	expectedPartitionResponse := make([]map[string]interface{}, 0)
	sampleParition := make(map[string]interface{})
	sampleParition["key"] = "region"
	sampleParition["value"] = "SEA"
	sampleParition["artifact_id"] = artifact.ArtifactID
	sampleParition["dataset_uuid"] = "uuid"
	expectedPartitionResponse = append(expectedPartitionResponse, sampleParition)
	return expectedPartitionResponse
}

// Raw db response to return on raw queries for tags
func getDBTagResponse(artifact models.Artifact) []map[string]interface{} {
	expectedTagResponse := make([]map[string]interface{}, 0)
	sampleTag := make(map[string]interface{})
	sampleTag["tag_name"] = "test-tag"
	sampleTag["artifact_id"] = artifact.ArtifactID
	sampleTag["dataset_uuid"] = "test-uuid"
	sampleTag["dataset_project"] = artifact.DatasetProject
	sampleTag["dataset_domain"] = artifact.DatasetDomain
	sampleTag["dataset_name"] = artifact.DatasetName
	sampleTag["dataset_version"] = artifact.DatasetVersion
	expectedTagResponse = append(expectedTagResponse, sampleTag)
	return expectedTagResponse
}

func TestCreateArtifact(t *testing.T) {
	artifact := getTestArtifact()

	artifactCreated := false
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	numArtifactDataCreated := 0
	numPartitionsCreated := 0

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "artifacts" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","dataset_uuid","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			artifactCreated = true
		},
	)

	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "artifact_data" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","name","location") VALUES (?,?,?,?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			numArtifactDataCreated++
		},
	)

	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "partitions" ("created_at","updated_at","deleted_at","dataset_uuid","key","value","artifact_id") VALUES (?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			numPartitionsCreated++
		},
	)

	data := make([]models.ArtifactData, 2)
	data[0] = models.ArtifactData{
		Name:     "test",
		Location: "dataloc",
	}
	data[1] = models.ArtifactData{
		Name:     "test2",
		Location: "dataloc2",
	}

	artifact.ArtifactData = data

	partitions := make([]models.Partition, 1)
	partitions[0] = getTestPartition()

	artifact.Partitions = partitions

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := artifactRepo.Create(context.Background(), artifact)
	assert.NoError(t, err)
	assert.True(t, artifactCreated)
	assert.Equal(t, 2, numArtifactDataCreated)
	assert.Equal(t, 1, numPartitionsCreated)
}

func TestGetArtifact(t *testing.T) {
	artifact := getTestArtifact()

	expectedArtifactDataResponse := getDBArtifactDataResponse(artifact)
	expectedArtifactResponse := getDBArtifactResponse(artifact)
	expectedPartitionResponse := getDBPartitionResponse(artifact)
	expectedTagResponse := getDBTagResponse(artifact)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL AND (("artifacts"."dataset_project" = testProject) AND ("artifacts"."dataset_name" = testName) AND ("artifacts"."dataset_domain" = testDomain) AND ("artifacts"."dataset_version" = testVersion) AND ("artifacts"."artifact_id" = 123)) ORDER BY artifacts.created_at DESC,"artifacts"."dataset_project" ASC LIMIT 1`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123)))) ORDER BY "artifact_data"."dataset_project" ASC`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions"  WHERE "partitions"."deleted_at" IS NULL AND (("artifact_id" IN (123))) ORDER BY "partitions"."dataset_uuid" ASC`).WithReply(expectedPartitionResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND ((("artifact_id","dataset_uuid") IN ((123,test-uuid)))) ORDER BY "tags"."dataset_project" ASC`).WithReply(expectedTagResponse)
	getInput := models.ArtifactKey{
		DatasetProject: artifact.DatasetProject,
		DatasetDomain:  artifact.DatasetDomain,
		DatasetName:    artifact.DatasetName,
		DatasetVersion: artifact.DatasetVersion,
		ArtifactID:     artifact.ArtifactID,
	}

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	response, err := artifactRepo.Get(context.Background(), getInput)
	assert.NoError(t, err)
	assert.Equal(t, artifact.ArtifactID, response.ArtifactID)
	assert.Equal(t, artifact.DatasetProject, response.DatasetProject)
	assert.Equal(t, artifact.DatasetDomain, response.DatasetDomain)
	assert.Equal(t, artifact.DatasetName, response.DatasetName)
	assert.Equal(t, artifact.DatasetVersion, response.DatasetVersion)

	assert.Equal(t, 1, len(response.ArtifactData))
	assert.Equal(t, 1, len(response.Partitions))
	assert.EqualValues(t, 1, len(response.Tags))
}

func TestGetArtifactByID(t *testing.T) {
	artifact := getTestArtifact()

	expectedArtifactDataResponse := getDBArtifactDataResponse(artifact)
	expectedArtifactResponse := getDBArtifactResponse(artifact)
	expectedPartitionResponse := getDBPartitionResponse(artifact)
	expectedTagResponse := getDBTagResponse(artifact)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL AND (("artifacts"."artifact_id" = 123)) ORDER BY artifacts.created_at DESC,"artifacts"."dataset_project" ASC LIMIT 1`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123)))) ORDER BY "artifact_data"."dataset_project" ASC`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions"  WHERE "partitions"."deleted_at" IS NULL AND (("artifact_id" IN (123))) ORDER BY "partitions"."dataset_uuid" ASC`).WithReply(expectedPartitionResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND ((("artifact_id","dataset_uuid") IN ((123,test-uuid)))) ORDER BY "tags"."dataset_project" ASC`).WithReply(expectedTagResponse)
	getInput := models.ArtifactKey{
		ArtifactID: artifact.ArtifactID,
	}

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	response, err := artifactRepo.Get(context.Background(), getInput)
	assert.NoError(t, err)
	assert.Equal(t, artifact.ArtifactID, response.ArtifactID)
}

func TestGetArtifactDoesNotExist(t *testing.T) {
	artifact := getTestArtifact()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	getInput := models.ArtifactKey{
		DatasetProject: artifact.DatasetProject,
		DatasetDomain:  artifact.DatasetDomain,
		DatasetName:    artifact.DatasetName,
		DatasetVersion: artifact.DatasetVersion,
		ArtifactID:     artifact.ArtifactID,
	}

	// by default mocket will return nil for any queries
	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	_, err := artifactRepo.Get(context.Background(), getInput)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.NotFound)
}

func TestCreateArtifactAlreadyExists(t *testing.T) {
	artifact := getTestArtifact()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "artifacts" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","dataset_uuid","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?,?)`).WithError(
		getAlreadyExistsErr(),
	)

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := artifactRepo.Create(context.Background(), artifact)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}

func TestListArtifactsWithPartition(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	dataset := getTestDataset()
	dataset.UUID = getDatasetUUID()

	artifact := getTestArtifact()
	expectedArtifactDataResponse := getDBArtifactDataResponse(artifact)
	expectedArtifactResponse := getDBArtifactResponse(artifact)
	expectedPartitionResponse := getDBPartitionResponse(artifact)
	expectedTagResponse := getDBTagResponse(artifact)
	GlobalMock.NewMock().WithQuery(
		`SELECT "artifacts".* FROM "artifacts" JOIN partitions partitions0 ON artifacts.artifact_id = partitions0.artifact_id WHERE "artifacts"."deleted_at" IS NULL AND ((partitions0.key = val1) AND (partitions0.val = val2) AND (artifacts.dataset_uuid = test-uuid)) ORDER BY artifacts.created_at desc LIMIT 10 OFFSET 10`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions"  WHERE "partitions"."deleted_at" IS NULL AND (("artifact_id" IN (123)))`).WithReply(expectedPartitionResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND ((("artifact_id","dataset_uuid") IN ((123,test-uuid))))`).WithReply(expectedTagResponse)

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	listInput := models.ListModelsInput{
		ModelFilters: []models.ModelFilter{
			{Entity: common.Partition,
				JoinCondition: NewGormJoinCondition(common.Artifact, common.Partition),
				ValueFilters: []models.ModelValueFilter{
					NewGormValueFilter(common.Equal, "key", "val1"),
					NewGormValueFilter(common.Equal, "val", "val2"),
				},
			},
		},
		Offset:        10,
		Limit:         10,
		SortParameter: NewGormSortParameter(datacatalog.PaginationOptions_CREATION_TIME, datacatalog.PaginationOptions_DESCENDING),
	}
	artifacts, err := artifactRepo.List(context.Background(), dataset.DatasetKey, listInput)
	assert.NoError(t, err)
	assert.Len(t, artifacts, 1)
	assert.Equal(t, artifacts[0].ArtifactID, artifact.ArtifactID)
	assert.Len(t, artifacts[0].ArtifactData, 1)
	assert.Len(t, artifacts[0].Partitions, 1)
	assert.Len(t, artifacts[0].Tags, 1)
}

func TestListArtifactsNoPartitions(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	dataset := getTestDataset()
	dataset.UUID = getDatasetUUID()

	artifact := getTestArtifact()
	expectedArtifactDataResponse := getDBArtifactDataResponse(artifact)
	expectedArtifactResponse := getDBArtifactResponse(artifact)
	expectedPartitionResponse := make([]map[string]interface{}, 0)

	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL AND ((artifacts.dataset_uuid = test-uuid)) LIMIT 10 OFFSET 10`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions"  WHERE "partitions"."deleted_at" IS NULL AND (("artifact_id" IN (123)))`).WithReply(expectedPartitionResponse)

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	listInput := models.ListModelsInput{
		Offset: 10,
		Limit:  10,
	}
	artifacts, err := artifactRepo.List(context.Background(), dataset.DatasetKey, listInput)
	assert.NoError(t, err)
	assert.Len(t, artifacts, 1)
	assert.Equal(t, artifacts[0].ArtifactID, artifact.ArtifactID)
	assert.Len(t, artifacts[0].ArtifactData, 1)
	assert.Len(t, artifacts[0].Partitions, 0)
}
