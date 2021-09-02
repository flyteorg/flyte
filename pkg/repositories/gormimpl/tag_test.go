package gormimpl

import (
	"testing"

	"github.com/jackc/pgconn"

	"gorm.io/gorm"

	"context"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	datacatalog_error "github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/utils"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"google.golang.org/grpc/codes"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

type pgError struct {
	e   error
	msg string
}

func (p *pgError) Error() string {
	return p.msg
}

func (p *pgError) Unwrap() error {
	return p.e
}

func getAlreadyExistsErr() error {
	return &pgError{
		e:   &pgconn.PgError{Code: "23505"},
		msg: "some error",
	}
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
		`INSERT INTO "tags" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","artifact_id","dataset_uuid") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`).WithCallback(
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
		`SELECT * FROM "tags" WHERE "tags"."dataset_project" = $1 AND "tags"."dataset_name" = $2 AND "tags"."dataset_domain" = $3 AND "tags"."dataset_version" = $4 AND "tags"."tag_name" = $5 ORDER BY tags.created_at DESC,"tags"."created_at" LIMIT 1%!!(string=test-tag)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(getDBTagResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifacts" WHERE ("artifacts"."dataset_project","artifacts"."dataset_name","artifacts"."dataset_domain","artifacts"."dataset_version","artifacts"."artifact_id") IN (($1,$2,$3,$4,$5))%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(getDBArtifactResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "artifact_data" WHERE ("artifact_data"."dataset_project","artifact_data"."dataset_name","artifact_data"."dataset_domain","artifact_data"."dataset_version","artifact_data"."artifact_id") IN (($1,$2,$3,$4,$5))%!!(string=123)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithReply(getDBArtifactDataResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "partitions" WHERE "partitions"."artifact_id" = $1 ORDER BY partitions.created_at ASC%!(EXTRA string=123)`).WithReply(getDBPartitionResponse(artifact))
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags" WHERE ("tags"."artifact_id","tags"."dataset_uuid") IN (($1,$2))%!!(string=test-uuid)(EXTRA string=123)`).WithReply(getDBTagResponse(artifact))
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

func TestTagNotFound(t *testing.T) {
	artifact := getTestArtifact()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tags" WHERE "tags"."dataset_project" = $1 AND "tags"."dataset_name" = $2 AND "tags"."dataset_domain" = $3 AND "tags"."dataset_version" = $4 AND "tags"."tag_name" = $5 ORDER BY tags.created_at DESC,"tags"."created_at" LIMIT 1%!!(string=test-tag)!(string=testVersion)!(string=testDomain)!(string=testName)(EXTRA string=testProject)`).WithError(
		gorm.ErrRecordNotFound,
	)
	getInput := models.TagKey{
		DatasetProject: artifact.DatasetProject,
		DatasetDomain:  artifact.DatasetDomain,
		DatasetName:    artifact.DatasetName,
		DatasetVersion: artifact.DatasetVersion,
		TagName:        "test-tag",
	}

	tagRepo := NewTagRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	_, err := tagRepo.Get(context.Background(), getInput)
	assert.Error(t, err)
	assert.Equal(t, "missing entity of type Tag with identifier ", err.Error())
}

func TestTagAlreadyExists(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`INSERT INTO "tags" ("created_at","updated_at","deleted_at","dataset_project","dataset_name","dataset_domain","dataset_version","tag_name","artifact_id","dataset_uuid") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`).WithError(
		getAlreadyExistsErr(),
	)

	tagRepo := NewTagRepo(utils.GetDbForTest(t), errors.NewPostgresErrorTransformer(), promutils.NewTestScope())
	err := tagRepo.Create(context.Background(), getTestTag())
	assert.Error(t, err)
	dcErr, ok := err.(datacatalog_error.DataCatalogError)
	assert.True(t, ok)
	assert.Equal(t, dcErr.Code().String(), codes.AlreadyExists.String())
}
