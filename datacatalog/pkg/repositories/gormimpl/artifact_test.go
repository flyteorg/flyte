package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	"github.com/flyteorg/datacatalog/pkg/common"
	apiErrors "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/utils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
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
		`INSERT INTO "artifacts" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","dataset_uuid","serialized_metadata") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			artifactCreated = true
		},
	)

	GlobalMock.NewMock().WithQuery(
		`INSERT INTO "artifact_data" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","name","location") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10),($11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ON CONFLICT ("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","name") DO UPDATE SET "dataset_project"="excluded"."dataset_project","dataset_name"="excluded"."dataset_name","dataset_domain"="excluded"."dataset_domain","dataset_version"="excluded"."dataset_version","artifact_id"="excluded"."artifact_id"`).WithCallback(
		func(s string, values []driver.NamedValue) {
			// Batch insert
			numArtifactDataCreated += 2
		},
	)

	GlobalMock.NewMock().WithQuery(
		`INSERT INTO "partitions" ("created_at","updated_at","deleted_at","dataset_uuid","key","value","artifact_id") VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT ("dataset_uuid","key","value","artifact_id") DO UPDATE SET "artifact_id"="excluded"."artifact_id"`).WithCallback(
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
		`SELECT * FROM "artifacts" WHERE "artifacts"."dataset_project" = $1 AND "artifacts"."dataset_name" = $2 AND "artifacts"."dataset_domain" = $3 AND "artifacts"."dataset_version" = $4 AND "artifacts"."artifact_id" = $5 ORDER BY artifacts.created_at DESC,"artifacts"."created_at" LIMIT 1%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data" WHERE ("artifact_data"."dataset_project","artifact_data"."dataset_name","artifact_data"."dataset_domain","artifact_data"."dataset_version","artifact_data"."artifact_id") IN (($1,$2,$3,$4,$5))%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions" WHERE "partitions"."artifact_id" = $1 ORDER BY partitions.created_at ASC%!(EXTRA string=123)`).WithReply(expectedPartitionResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags" WHERE ("tags"."artifact_id","tags"."dataset_uuid") IN (($1,$2))%!!(string=test-uuid)(EXTRA string=123)`).WithReply(expectedTagResponse)
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
		`SELECT * FROM "artifacts" WHERE "artifacts"."artifact_id" = $1 ORDER BY artifacts.created_at DESC,"artifacts"."created_at" LIMIT 1%!(EXTRA string=123)`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data" WHERE ("artifact_data"."dataset_project","artifact_data"."dataset_name","artifact_data"."dataset_domain","artifact_data"."dataset_version","artifact_data"."artifact_id") IN (($1,$2,$3,$4,$5))%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions" WHERE "partitions"."artifact_id" = $1 ORDER BY partitions.created_at ASC%!(EXTRA string=123)`).WithReply(expectedPartitionResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags" WHERE ("tags"."artifact_id","tags"."dataset_uuid") IN (($1,$2))%!!(string=test-uuid)(EXTRA string=123)`).WithReply(expectedTagResponse)
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
		`INSERT INTO "artifacts" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","dataset_uuid","serialized_metadata") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`).WithError(
		getAlreadyExistsErr(),
	)

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := artifactRepo.Create(context.Background(), artifact)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code().String(), codes.AlreadyExists.String())
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
		`SELECT "artifacts"."created_at","artifacts"."updated_at","artifacts"."deleted_at","artifacts"."dataset_project","artifacts"."dataset_name","artifacts"."dataset_domain","artifacts"."dataset_version","artifacts"."artifact_id","artifacts"."dataset_uuid","artifacts"."serialized_metadata" FROM "artifacts" JOIN partitions partitions0 ON artifacts.artifact_id = partitions0.artifact_id WHERE partitions0.key = $1 AND partitions0.val = $2 AND artifacts.dataset_uuid = $3 ORDER BY artifacts.created_at desc LIMIT 10 OFFSET 10%!!(string=test-uuid)!(string=val2)(EXTRA string=val1)`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data" WHERE ("artifact_data"."dataset_project","artifact_data"."dataset_name","artifact_data"."dataset_domain","artifact_data"."dataset_version","artifact_data"."artifact_id") IN (($1,$2,$3,$4,$5))%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions" WHERE "partitions"."artifact_id" = $1 ORDER BY partitions.created_at ASC%!(EXTRA string=123)`).WithReply(expectedPartitionResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags" WHERE ("tags"."artifact_id","tags"."dataset_uuid") IN (($1,$2))%!!(string=test-uuid)(EXTRA string=123)`).WithReply(expectedTagResponse)

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
		`SELECT * FROM "artifacts" WHERE artifacts.dataset_uuid = $1 LIMIT 10 OFFSET 10%!(EXTRA string=test-uuid)`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data" WHERE ("artifact_data"."dataset_project","artifact_data"."dataset_name","artifact_data"."dataset_domain","artifact_data"."dataset_version","artifact_data"."artifact_id") IN (($1,$2,$3,$4,$5))%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(expectedArtifactDataResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions" WHERE "partitions"."artifact_id" = $1 ORDER BY partitions.created_at ASC%!(EXTRA string=123)`).WithReply(expectedPartitionResponse)

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

func TestUpdateArtifact(t *testing.T) {
	ctx := context.Background()
	artifact := getTestArtifact()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	artifactUpdated := false
	GlobalMock.NewMock().WithQuery(`UPDATE "artifacts" SET "updated_at"=$1,"artifact_id"=$2 WHERE "artifact_id" = $3`).
		WithRowsNum(1).
		WithCallback(func(s string, values []driver.NamedValue) {
			artifactUpdated = true
		})
	artifactDataDeleted := false
	GlobalMock.NewMock().
		WithQuery(`DELETE FROM "artifact_data" WHERE "artifact_data"."artifact_id" = $1 AND name NOT IN ($2,$3)`).
		WithRowsNum(0).
		WithCallback(func(s string, values []driver.NamedValue) {
			artifactDataDeleted = true
		})
	artifactDataUpserted := false
	GlobalMock.NewMock().WithQuery(`INSERT INTO "artifact_data" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","name","location") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10),($11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ON CONFLICT DO NOTHING`).
		WithRowsNum(1).
		WithCallback(func(s string, values []driver.NamedValue) {
			artifactDataUpserted = true
		})

	updateInput := models.Artifact{
		ArtifactKey: models.ArtifactKey{
			ArtifactID: artifact.ArtifactID,
		},
		ArtifactData: []models.ArtifactData{
			{
				Name:     "test-dataloc-name",
				Location: "test-dataloc-location",
			},
			{
				Name:     "additional-test-dataloc-name",
				Location: "additional-test-dataloc-location",
			},
		},
	}

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := artifactRepo.Update(ctx, updateInput)
	assert.NoError(t, err)
	assert.True(t, artifactUpdated)
	assert.True(t, artifactDataDeleted)
	assert.True(t, artifactDataUpserted)
}

func TestUpdateArtifactDoesNotExist(t *testing.T) {
	ctx := context.Background()
	artifact := getTestArtifact()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	updateInput := models.Artifact{
		ArtifactKey: models.ArtifactKey{
			ArtifactID: artifact.ArtifactID,
		},
		ArtifactData: []models.ArtifactData{
			{
				Name:     "test-dataloc-name",
				Location: "test-dataloc-location",
			},
			{
				Name:     "additional-test-dataloc-name",
				Location: "additional-test-dataloc-location",
			},
		},
	}

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := artifactRepo.Update(ctx, updateInput)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.NotFound)
}

func TestUpdateArtifactError(t *testing.T) {
	artifact := getTestArtifact()

	t.Run("ArtifactUpdate", func(t *testing.T) {
		ctx := context.Background()

		GlobalMock := mocket.Catcher.Reset()
		GlobalMock.Logging = true

		GlobalMock.NewMock().WithQuery(`UPDATE "artifacts" SET "updated_at"=$1,"artifact_id"=$2 WHERE "artifact_id" = $3`).
			WithExecException()

		updateInput := models.Artifact{
			ArtifactKey: models.ArtifactKey{
				ArtifactID: artifact.ArtifactID,
			},
			ArtifactData: []models.ArtifactData{
				{
					Name:     "test-dataloc-name",
					Location: "test-dataloc-location",
				},
				{
					Name:     "additional-test-dataloc-name",
					Location: "additional-test-dataloc-location",
				},
			},
		}

		artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
		err := artifactRepo.Update(ctx, updateInput)
		assert.Error(t, err)
		dcErr, ok := err.(apiErrors.DataCatalogError)
		assert.True(t, ok)
		assert.Equal(t, dcErr.Code(), codes.Internal)
	})

	t.Run("ArtifactDataDelete", func(t *testing.T) {
		ctx := context.Background()

		GlobalMock := mocket.Catcher.Reset()
		GlobalMock.Logging = true

		artifactUpdated := false
		GlobalMock.NewMock().WithQuery(`UPDATE "artifacts" SET "updated_at"=$1,"artifact_id"=$2 WHERE "artifact_id" = $3`).
			WithRowsNum(1).
			WithCallback(func(s string, values []driver.NamedValue) {
				artifactUpdated = true
			})
		GlobalMock.NewMock().
			WithQuery(`DELETE FROM "artifact_data" WHERE "artifact_data"."artifact_id" = $1 AND name NOT IN ($2,$3)`).
			WithExecException()

		updateInput := models.Artifact{
			ArtifactKey: models.ArtifactKey{
				ArtifactID: artifact.ArtifactID,
			},
			ArtifactData: []models.ArtifactData{
				{
					Name:     "test-dataloc-name",
					Location: "test-dataloc-location",
				},
				{
					Name:     "additional-test-dataloc-name",
					Location: "additional-test-dataloc-location",
				},
			},
		}

		artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
		err := artifactRepo.Update(ctx, updateInput)
		assert.Error(t, err)
		dcErr, ok := err.(apiErrors.DataCatalogError)
		assert.True(t, ok)
		assert.Equal(t, dcErr.Code(), codes.Internal)
		assert.True(t, artifactUpdated)
	})

	t.Run("ArtifactDataUpsert", func(t *testing.T) {
		ctx := context.Background()

		GlobalMock := mocket.Catcher.Reset()
		GlobalMock.Logging = true

		artifactUpdated := false
		GlobalMock.NewMock().WithQuery(`UPDATE "artifacts" SET "updated_at"=$1,"artifact_id"=$2 WHERE "artifact_id" = $3`).
			WithRowsNum(1).
			WithCallback(func(s string, values []driver.NamedValue) {
				artifactUpdated = true
			})
		artifactDataDeleted := false
		GlobalMock.NewMock().
			WithQuery(`DELETE FROM "artifact_data" WHERE "artifact_data"."artifact_id" = $1 AND name NOT IN ($2,$3)`).
			WithRowsNum(0).
			WithCallback(func(s string, values []driver.NamedValue) {
				artifactDataDeleted = true
			})
		GlobalMock.NewMock().WithQuery(`INSERT INTO "artifact_data" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","name","location") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10),($11,$12,$13,$14,$15,$16,$17,$18,$19,$20) ON CONFLICT DO NOTHING`).
			WithExecException()

		updateInput := models.Artifact{
			ArtifactKey: models.ArtifactKey{
				ArtifactID: artifact.ArtifactID,
			},
			ArtifactData: []models.ArtifactData{
				{
					Name:     "test-dataloc-name",
					Location: "test-dataloc-location",
				},
				{
					Name:     "additional-test-dataloc-name",
					Location: "additional-test-dataloc-location",
				},
			},
		}

		artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
		err := artifactRepo.Update(ctx, updateInput)
		assert.Error(t, err)
		dcErr, ok := err.(apiErrors.DataCatalogError)
		assert.True(t, ok)
		assert.Equal(t, dcErr.Code(), codes.Internal)
		assert.True(t, artifactUpdated)
		assert.True(t, artifactDataDeleted)
	})
}
