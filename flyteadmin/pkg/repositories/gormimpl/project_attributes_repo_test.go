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

const testProjectAttr = "project"
const testResourceAttr = "resource"

func TestCreateProjectAttributes(t *testing.T) {
	projectRepo := NewProjectAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT INTO "project_attributes" ` +
			`("created_at","updated_at","deleted_at","project","resource","attributes") VALUES (?,?,?,?,?,?)`)

	err := projectRepo.CreateOrUpdate(context.Background(), models.ProjectAttributes{
		Project:    testProjectAttr,
		Resource:   testResourceAttr,
		Attributes: []byte("attrs"),
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetProjectAttributes(t *testing.T) {
	projectRepo := NewProjectAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["project"] = testProjectAttr
	response["resource"] = testResourceAttr
	response["attributes"] = []byte("attrs")

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "project_attributes"  WHERE "project_attributes"."deleted_at" ` +
		`IS NULL AND (("project_attributes"."project" = project) AND ("project_attributes"."resource" = resource)) ` +
		`ORDER BY "project_attributes"."id" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := projectRepo.Get(context.Background(), "project", "resource")
	assert.Nil(t, err)
	assert.Equal(t, testProjectAttr, output.Project)
	assert.Equal(t, testResourceAttr, output.Resource)
	assert.Equal(t, []byte("attrs"), output.Attributes)
}

func TestDeleteProjectAttributes(t *testing.T) {
	projectRepo := NewProjectAttributesRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	fakeResponse := query.WithQuery(
		`DELETE FROM "project_attributes"  WHERE ("project_attributes"."project" = ?) AND ` +
			`("project_attributes"."resource" = ?)`)

	err := projectRepo.Delete(context.Background(), "project", "resource")
	assert.Nil(t, err)
	assert.True(t, fakeResponse.Triggered)
}
