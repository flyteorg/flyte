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

func TestCreateProjectDomainAttributes(t *testing.T) {
	projectDomainRepo := NewProjectDomainAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT INTO "project_domain_attributes" ` +
			`("created_at","updated_at","deleted_at","project","domain","resource","attributes") VALUES (?,?,?,?,?,?,?)`)

	err := projectDomainRepo.CreateOrUpdate(context.Background(), models.ProjectDomainAttributes{
		Project:    "project",
		Domain:     "domain",
		Resource:   "resource",
		Attributes: []byte("attrs"),
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetProjectDomainAttributes(t *testing.T) {
	projectDomainRepo := NewProjectDomainAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["project"] = "project"
	response["domain"] = "domain"
	response["resource"] = "resource"
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "project_domain_attributes"  WHERE "project_domain_attributes"."deleted_at" ` +
		`IS NULL AND (("project_domain_attributes"."project" = project) AND ("project_domain_attributes"."domain" = ` +
		`domain) AND ("project_domain_attributes"."resource" = resource)) ORDER BY "project_domain_attributes"."id" ` +
		`ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := projectDomainRepo.Get(context.Background(), "project", "domain", "resource")
	assert.Nil(t, err)
	assert.Equal(t, "project", output.Project)
	assert.Equal(t, "domain", output.Domain)
	assert.Equal(t, "resource", output.Resource)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestDeleteProjectDomainAttributes(t *testing.T) {
	projectDomainRepo := NewProjectDomainAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	fakeResponse := query.WithQuery(
		`DELETE FROM "project_domain_attributes"  WHERE ("project_domain_attributes"."project" = ?) AND ` +
			`("project_domain_attributes"."domain" = ?) AND ("project_domain_attributes"."resource" = ?)`)

	err := projectDomainRepo.Delete(context.Background(), "project", "domain", "resource")
	assert.Nil(t, err)
	assert.True(t, fakeResponse.Triggered)
}
