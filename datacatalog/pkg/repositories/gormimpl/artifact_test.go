package gormimpl

import (
	"testing"

	"context"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	apiErrors "github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
	"google.golang.org/grpc/codes"
)

func getTestArtifact() models.Artifact {
	return models.Artifact{
		ArtifactKey: models.ArtifactKey{
			ArtifactID:     "123",
			DatasetProject: "testProject",
			DatasetDomain:  "testDomain",
			DatasetName:    "testName",
			DatasetVersion: "testVersion",
		},
	}
}

func TestCreateArtifact(t *testing.T) {
	artifact := getTestArtifact()

	artifactCreated := false
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	numArtifactDataCreated := 0

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "artifacts" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?)`).WithCallback(
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

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	err := artifactRepo.Create(context.Background(), artifact)
	assert.NoError(t, err)
	assert.True(t, artifactCreated)
	assert.Equal(t, 2, numArtifactDataCreated)
}

func TestGetArtifact(t *testing.T) {
	artifact := getTestArtifact()

	expectedArtifactDataResponse := make([]map[string]interface{}, 0)
	sampleArtifactData := make(map[string]interface{})
	sampleArtifactData["dataset_project"] = artifact.DatasetProject
	sampleArtifactData["dataset_domain"] = artifact.DatasetDomain
	sampleArtifactData["dataset_name"] = artifact.DatasetName
	sampleArtifactData["dataset_version"] = artifact.DatasetVersion
	sampleArtifactData["artifact_id"] = artifact.ArtifactID
	sampleArtifactData["name"] = "test-dataloc-name"
	sampleArtifactData["location"] = "test-dataloc-location"

	expectedArtifactDataResponse = append(expectedArtifactDataResponse, sampleArtifactData)

	expectedArtifactResponse := make([]map[string]interface{}, 0)
	sampleArtifact := make(map[string]interface{})
	sampleArtifact["dataset_project"] = artifact.DatasetProject
	sampleArtifact["dataset_domain"] = artifact.DatasetDomain
	sampleArtifact["dataset_name"] = artifact.DatasetName
	sampleArtifact["dataset_version"] = artifact.DatasetVersion
	sampleArtifact["artifact_id"] = artifact.ArtifactID
	expectedArtifactResponse = append(expectedArtifactResponse, sampleArtifact)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL AND (("artifacts"."dataset_project" = testProject) AND ("artifacts"."dataset_name" = testName) AND ("artifacts"."dataset_domain" = testDomain) AND ("artifacts"."dataset_version" = testVersion) AND ("artifacts"."artifact_id" = 123))`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(expectedArtifactDataResponse)
	getInput := models.ArtifactKey{
		DatasetProject: artifact.DatasetProject,
		DatasetDomain:  artifact.DatasetDomain,
		DatasetName:    artifact.DatasetName,
		DatasetVersion: artifact.DatasetVersion,
		ArtifactID:     artifact.ArtifactID,
	}

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	response, err := artifactRepo.Get(context.Background(), getInput)
	assert.NoError(t, err)
	assert.Equal(t, artifact.ArtifactID, response.ArtifactID)
	assert.Equal(t, artifact.DatasetProject, response.DatasetProject)
	assert.Equal(t, artifact.DatasetDomain, response.DatasetDomain)
	assert.Equal(t, artifact.DatasetName, response.DatasetName)
	assert.Equal(t, artifact.DatasetVersion, response.DatasetVersion)

	assert.Equal(t, 1, len(response.ArtifactData))
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
	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
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
		`INSERT  INTO "artifacts" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id","serialized_metadata") VALUES (?,?,?,?,?,?,?,?,?)`).WithError(
		getAlreadyExistsErr(),
	)

	artifactRepo := NewArtifactRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer())
	err := artifactRepo.Create(context.Background(), artifact)
	assert.Error(t, err)
	dcErr, ok := err.(apiErrors.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}
