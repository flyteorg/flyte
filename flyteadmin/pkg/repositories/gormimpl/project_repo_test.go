package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

var alphabeticalSortParam, _ = common.NewSortParameter(&admin.Sort{
	Direction: admin.Sort_ASCENDING,
	Key:       "identifier",
}, models.ProjectColumns)

func TestCreateProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	GlobalMock.Logging = true
	query.WithQuery(
		`INSERT INTO "projects" ("created_at","updated_at","deleted_at","name","description","labels","state","identifier") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)

	activeState := int32(admin.Project_ACTIVE)
	err := projectRepo.Create(context.Background(), models.Project{
		Identifier:  "proj",
		Name:        "proj",
		Description: "projDescription",
		State:       &activeState,
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
	response["state"] = admin.Project_ACTIVE

	output, err := projectRepo.Get(context.Background(), "project_id")
	assert.Empty(t, output)
	assert.EqualError(t, err, "project [project_id] not found")

	query := GlobalMock.NewMock()
	GlobalMock.Logging = true
	query.WithQuery(`SELECT * FROM "projects" WHERE "projects"."identifier" = $1 LIMIT 1`).WithReply(
		[]map[string]interface{}{
			response,
		})

	output, err = projectRepo.Get(context.Background(), "project_id")
	assert.Nil(t, err)
	assert.Equal(t, "project_id", output.Identifier)
	assert.Equal(t, "project_name", output.Name)
	assert.Equal(t, "project_description", output.Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output.State)
}

func testListProjects(input interfaces.ListResourceInput, sql string, t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	projects := make([]map[string]interface{}, 1)
	fooProject := make(map[string]interface{})
	fooProject["identifier"] = "foo"
	fooProject["name"] = "foo =)"
	fooProject["description"] = "foo description"
	fooProject["state"] = admin.Project_ACTIVE
	projects[0] = fooProject

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(sql).
		WithReply(projects)

	output, err := projectRepo.List(context.Background(), input)
	assert.Nil(t, err)
	assert.Len(t, output, 1)
	assert.Equal(t, "foo", output[0].Identifier)
	assert.Equal(t, "foo =)", output[0].Name)
	assert.Equal(t, "foo description", output[0].Description)
	assert.Equal(t, int32(admin.Project_ACTIVE), *output[0].State)
}

func TestListProjects(t *testing.T) {
	filter, err := common.NewSingleValueFilter(common.Project, common.Equal, "name", "foo")

	assert.NoError(t, err)
	testListProjects(interfaces.ListResourceInput{
		Offset:        0,
		Limit:         1,
		InlineFilters: []common.InlineFilter{filter},
		SortParameter: alphabeticalSortParam,
	}, `SELECT * FROM "projects" WHERE name = $1 ORDER BY identifier asc LIMIT 1`, t)
}

func TestListProjects_NoFilters(t *testing.T) {
	testListProjects(interfaces.ListResourceInput{
		Offset:        0,
		Limit:         1,
		SortParameter: alphabeticalSortParam,
	}, `SELECT * FROM "projects" WHERE state != $1 ORDER BY identifier asc`, t)
}

func TestListProjects_NoLimit(t *testing.T) {
	testListProjects(interfaces.ListResourceInput{
		Offset:        0,
		SortParameter: alphabeticalSortParam,
	}, `SELECT * FROM "projects" WHERE state != $1 ORDER BY identifier asc`, t)
}

func TestUpdateProject(t *testing.T) {
	projectRepo := NewProjectRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	query := GlobalMock.NewMock()
	GlobalMock.Logging = true
	query.WithQuery(`UPDATE "projects" SET "updated_at"=$1,"identifier"=$2,"name"=$3,"description"=$4,"state"=$5 WHERE "identifier" = $6`)

	activeState := int32(admin.Project_ACTIVE)
	err := projectRepo.UpdateProject(context.Background(), models.Project{
		Identifier:  "project_id",
		Name:        "project_name",
		Description: "project_description",
		State:       &activeState,
	})
	assert.Nil(t, err)
	assert.True(t, query.Triggered)
}
