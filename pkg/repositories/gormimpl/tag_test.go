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

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND (("tags"."dataset_project" = testProject) AND ("tags"."dataset_name" = testName) AND ("tags"."dataset_domain" = testDomain) AND ("tags"."dataset_version" = testVersion) AND ("tags"."tag_name" = test-tag)) ORDER BY tags.created_at DESC,"tags"."dataset_project" ASC LIMIT 1`).WithReply(getDBTagResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts"  WHERE "artifacts"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(getDBArtifactResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data"  WHERE "artifact_data"."deleted_at" IS NULL AND ((("dataset_project","dataset_name","dataset_domain","dataset_version","artifact_id") IN ((testProject,testName,testDomain,testVersion,123))))`).WithReply(getDBArtifactDataResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions"  WHERE "partitions"."deleted_at" IS NULL AND (("artifact_id" IN (123)))`).WithReply(getDBPartitionResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags"  WHERE "tags"."deleted_at" IS NULL AND ((("artifact_id","dataset_uuid") IN ((123,test-uuid))))`).WithReply(getDBTagResponse(artifact))
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
	assert.Len(t, response.Artifact.Partitions, 1)
	assert.Len(t, response.Artifact.Tags, 1)
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
