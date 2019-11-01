package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectDomain(t *testing.T) {
	projectRepo := NewProjectDomainRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT  INTO "project_domains" ` +
			`("created_at","updated_at","deleted_at","project","domain","attributes") VALUES (?,?,?,?,?,?)`)

	err := projectRepo.CreateOrUpdate(context.Background(), models.ProjectDomain{
		Project:    "project",
		Domain:     "domain",
		Attributes: []byte("attrs"),
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetProjectDomain(t *testing.T) {
	projectRepo := NewProjectDomainRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["project"] = "project"
	response["domain"] = "domain"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "project_domains"  WHERE "project_domains"."deleted_at" IS NULL AND ` +
		`(("project_domains"."project" = project) AND ("project_domains"."domain" = domain)) ORDER BY ` +
		`"project_domains"."id" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := projectRepo.Get(context.Background(), "project", "domain")
	assert.Nil(t, err)
	assert.Equal(t, "project", output.Project)
	assert.Equal(t, "domain", output.Domain)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}
