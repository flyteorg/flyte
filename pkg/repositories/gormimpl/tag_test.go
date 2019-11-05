package gormimpl

import (
	"testing"

	"context"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	datacatalog_error "github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/repositories/errors"

	"github.com/lib/pq"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/utils"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"google.golang.org/grpc/codes"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func getAlreadyExistsErr() error {
	return &pq.Error{Code: "23505"}
}

func getTestTag() models.Tag {
	artifact := getTestArtifact()
	return models.Tag{
		TagKey: models.TagKey{
			DatasetProject: artifact.DatasetProject,
			DatasetDomain:  artifact.DatasetDomain,
			DatasetName:    artifact.DatasetName,
			DatasetVersion: artifact.DatasetVersion,
			TagName:        "test-tagname",
		},
		ArtifactID: artifact.ArtifactID,
	}
}

func TestCreateTag(t *testing.T) {
	tagCreated := false
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "tags" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","artifact_id","dataset_uuid") VALUES (?,?,?,?,?,?,?,?,?,?)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			tagCreated = true
		},
	)

	tagRepo := NewTagRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := tagRepo.Create(context.Background(), getTestTag())
	assert.NoError(t, err)
	assert.True(t, tagCreated)
}

func TestGetTag(t *testing.T) {
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
	sampleArtifactData["dataset_uuid"] = artifact.DatasetUUID

	expectedArtifactDataResponse = append(expectedArtifactDataResponse, sampleArtifactData)

	expectedArtifactResponse := make([]map[string]interface{}, 0)
	sampleArtifact := make(map[string]interface{})
	sampleArtifact["dataset_project"] = artifact.DatasetProject
	sampleArtifact["dataset_domain"] = artifact.DatasetDomain
	sampleArtifact["dataset_name"] = artifact.DatasetName
	sampleArtifact["dataset_version"] = artifact.DatasetVersion
	sampleArtifact["artifact_id"] = artifact.ArtifactID
	sampleArtifact["dataset_uuid"] = artifact.DatasetUUID
	expectedArtifactResponse = append(expectedArtifactResponse, sampleArtifact)

	expectedTagResponse := make([]map[string]interface{}, 0)
	sampleTag := make(map[string]interface{})
	sampleTag["dataset_project"] = artifact.DatasetProject
	sampleTag["dataset_domain"] = artifact.DatasetDomain
	sampleTag["dataset_name"] = artifact.DatasetName
	sampleTag["dataset_version"] = artifact.DatasetVersion
	sampleTag["artifact_id"] = artifact.ArtifactID
	sampleTag["name"] = "test-tag"
	sampleTag["dataset_uuid"] = artifact.DatasetUUID
	expectedTagResponse = append(expectedTagResponse, sampleTag)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND (("tags"."dataset_project" = testProject) AND ("tags"."dataset_name" = testName) AND ("tags"."dataset_domain" = testDomain) AND ("tags"."dataset_version" = testVersion) AND ("tags"."tag_name" = test-tag))`).WithReply(expectedTagResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(expectedArtifactResponse)
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(expectedArtifactDataResponse)

	getInput := models.TagKey{
		DatasetProject: artifact.DatasetProject,
		DatasetDomain:  artifact.DatasetDomain,
		DatasetName:    artifact.DatasetName,
		DatasetVersion: artifact.DatasetVersion,
		TagName:        "test-tag",
	}

	tagRepo := NewTagRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	response, err := tagRepo.Get(context.Background(), getInput)
	assert.NoError(t, err)
	assert.Equal(t, artifact.ArtifactID, response.ArtifactID)
	assert.Equal(t, artifact.ArtifactID, response.Artifact.ArtifactID)
	assert.Len(t, response.Artifact.ArtifactData, 1)
}

func TestTagAlreadyExists(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT  INTO "tags" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","artifact_id","dataset_uuid") VALUES (?,?,?,?,?,?,?,?,?,?)`).WithError(
		getAlreadyExistsErr(),
	)

	tagRepo := NewTagRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := tagRepo.Create(context.Background(), getTestTag())
	assert.Error(t, err)
	dcErr, ok := err.(datacatalog_error.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code(), codes.AlreadyExists)
}
