package gormimpl

import (
	"context"
	"testing"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

const pythonTestTaskType = "python-task"

func TestCreateTask(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := taskRepo.Create(context.Background(), models.Task{
		TaskKey: models.TaskKey{
			Project: project,
			Domain:  domain,
			Name:    name,
			Version: version,
		},
		Closure: []byte{1, 2},
		Type:    pythonTestTaskType,
	}, &models.DescriptionEntity{ShortDescription: "hello"})
	assert.NoError(t, err)
}

func getMockTaskResponseFromDb(version string, spec []byte) map[string]interface{} {
	task := make(map[string]interface{})
	task["project"] = project
	task["domain"] = domain
	task["name"] = name
	task["version"] = version
	task["closure"] = spec
	task["type"] = pythonTestTaskType
	return task
}

func TestGetTask(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	tasks := make([]map[string]interface{}, 0)
	task := getMockTaskResponseFromDb(version, []byte{1, 2})
	tasks = append(tasks, task)

	output, err := taskRepo.Get(context.Background(), interfaces.Identifier{
		Project: project,
		Domain:  domain,
		Name:    name,
		Version: version,
	})
	assert.Empty(t, output)
	assert.EqualError(t, err, "missing entity of type TASK with identifier project:\"project\" domain:\"domain\" name:\"name\" version:\"XYZ\" ")

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "tasks" WHERE "tasks"."project" = $1 AND "tasks"."domain" = $2 AND "tasks"."name" = $3 AND "tasks"."version" = $4 LIMIT 1`).
		WithReply(tasks)
	output, err = taskRepo.Get(context.Background(), interfaces.Identifier{
		Project: project,
		Domain:  domain,
		Name:    name,
		Version: version,
	})
	assert.Empty(t, err)
	assert.Equal(t, project, output.Project)
	assert.Equal(t, domain, output.Domain)
	assert.Equal(t, name, output.Name)
	assert.Equal(t, version, output.Version)
	assert.Equal(t, []byte{1, 2}, output.Closure)
	assert.Equal(t, pythonTestTaskType, output.Type)
}

func TestListTasks(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	tasks := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "XYZ"}
	spec := []byte{1, 2}
	for _, version := range versions {
		task := getMockTaskResponseFromDb(version, spec)
		tasks = append(tasks, task)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(tasks)

	collection, err := taskRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
		},
		Limit: 20,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Tasks)
	assert.Len(t, collection.Tasks, 2)
	for _, task := range collection.Tasks {
		assert.Equal(t, project, task.Project)
		assert.Equal(t, domain, task.Domain)
		assert.Equal(t, name, task.Name)
		assert.Contains(t, versions, task.Version)
		assert.Equal(t, spec, task.Closure)
		assert.Equal(t, pythonTestTaskType, task.Type)
	}
}

func TestListTasks_Pagination(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	tasks := make([]map[string]interface{}, 0)
	versions := []string{"ABC", "DEF"}
	spec := []byte{1, 2}
	for _, version := range versions {
		task := getMockTaskResponseFromDb(version, spec)
		tasks = append(tasks, task)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(tasks)

	collection, err := taskRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
		},
		Limit: 2,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Tasks)
	assert.Len(t, collection.Tasks, 2)
	for idx, task := range collection.Tasks {
		assert.Equal(t, project, task.Project)
		assert.Equal(t, domain, task.Domain)
		assert.Equal(t, name, task.Name)
		assert.Equal(t, versions[idx], task.Version)
		assert.Equal(t, spec, task.Closure)
		assert.Equal(t, pythonTestTaskType, task.Type)
	}
}

func TestListTasks_Filters(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	tasks := make([]map[string]interface{}, 0)
	task := getMockTaskResponseFromDb("ABC", []byte{1, 2})
	tasks = append(tasks, task)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append the name filter
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "tasks" WHERE project = $1 AND domain = $2 AND name = $3 AND version = $4 LIMIT 20`).WithReply(tasks[0:1])

	collection, err := taskRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
			getEqualityFilter(common.Task, "version", "ABC"),
		},
		Limit: 20,
	})
	assert.Empty(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Tasks)
	assert.Len(t, collection.Tasks, 1)
	assert.Equal(t, project, collection.Tasks[0].Project)
	assert.Equal(t, domain, collection.Tasks[0].Domain)
	assert.Equal(t, name, collection.Tasks[0].Name)
	assert.Equal(t, "ABC", collection.Tasks[0].Version)
	assert.Equal(t, []byte{1, 2}, collection.Tasks[0].Closure)
	assert.Equal(t, pythonTestTaskType, collection.Tasks[0].Type)
}

func TestListTasks_Order(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	tasks := make([]map[string]interface{}, 0)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that include ordering by project
	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(`project desc`)
	mockQuery.WithReply(tasks)

	sortParameter, _ := common.NewSortParameter(&admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	}, models.TaskColumns)
	_, err := taskRepo.List(context.Background(), interfaces.ListResourceInput{
		SortParameter: sortParameter,
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
			getEqualityFilter(common.Task, "version", "ABC"),
		},
		Limit: 20,
	})
	assert.Empty(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListTasks_MissingParameters(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := taskRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
			getEqualityFilter(common.Task, "version", version),
		},
	})
	assert.Equal(t, err.Error(), "missing and/or invalid parameters: limit")

	_, err = taskRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit: 20,
	})
	assert.Equal(t, err.Error(), "missing and/or invalid parameters: filters")
}

func TestListTaskIds(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	tasks := make([]map[string]interface{}, 0)
	versions := []string{"v3", "v4"}
	spec := []byte{1, 2}
	for _, version := range versions {
		task := getMockTaskResponseFromDb(version, spec)
		tasks = append(tasks, task)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(`GROUP BY project, domain, name`).WithReply(tasks)

	collection, err := taskRepo.ListTaskIdentifiers(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Tasks)
	assert.Len(t, collection.Tasks, 2)
	for idx, task := range collection.Tasks {
		assert.Equal(t, project, task.Project)
		assert.Equal(t, domain, task.Domain)
		assert.Equal(t, name, task.Name)
		assert.Equal(t, versions[idx], task.Version)
		assert.Equal(t, spec, task.Closure)
		assert.Equal(t, pythonTestTaskType, task.Type)
	}
}

func TestListTaskIds_MissingParameters(t *testing.T) {
	taskRepo := NewTaskRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	_, err := taskRepo.ListTaskIdentifiers(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
		},
	})

	// Limit must be specified
	assert.Equal(t, "missing and/or invalid parameters: limit", err.Error())
}
