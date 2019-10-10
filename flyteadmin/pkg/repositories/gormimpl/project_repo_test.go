package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	query.WithQuery(
		`INSERT  INTO "projects" ("created_at","updated_at","deleted_at","identifier","name","description") VALUES (?,?,?,?,?,?)`)

	err := projectRepo.Create(context.Background(), models.Project{
		Identifier:  "proj",
		Name:        "proj",
		Description: "projDescription",
	})
	assert.NoError(t, err)
	assert.True(t, query.Triggered)
}

func TestGetProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	response := make(map[string]interface{})
	response["identifier"] = "project_id"
	response["name"] = "project_name"
	response["description"] = "project_description"

	query := GlobalMock.NewMock()
	query.WithQuery(`SELECT * FROM "projects"  WHERE "projects"."deleted_at" IS NULL AND ` +
		`(("projects"."identifier" = project_id)) ORDER BY "projects"."identifier" ASC LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err := projectRepo.Get(context.Background(), "project_id")
	assert.Nil(t, err)
	assert.Equal(t, "project_id", output.Identifier)
	assert.Equal(t, "project_name", output.Name)
	assert.Equal(t, "project_description", output.Description)
}

func TestListProjects(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	projects := make([]map[string]interface{}, 2)
	fooProject := make(map[string]interface{})
	fooProject["identifier"] = "foo"
	fooProject["name"] = "foo =)"
	fooProject["description"] = "foo description"
	projects[0] = fooProject

	barProject := make(map[string]interface{})
	barProject["identifier"] = "bar"
	barProject["name"] = "Bar"
	barProject["description"] = "Bar description"
	projects[1] = barProject

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "projects"  WHERE "projects"."deleted_at" IS NULL ORDER BY identifier asc`).
		WithReply(projects)

	var alphabeticalSortParam, _ = common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "identifier",
	})
	output, err := projectRepo.ListAll(context.Background(), alphabeticalSortParam)
	assert.Nil(t, err)
	assert.Len(t, output, 2)
	assert.Equal(t, "foo", output[0].Identifier)
	assert.Equal(t, "foo =)", output[0].Name)
	assert.Equal(t, "foo description", output[0].Description)
	assert.Equal(t, "bar", output[1].Identifier)
	assert.Equal(t, "Bar", output[1].Name)
	assert.Equal(t, "Bar description", output[1].Description)
}
