package gormimpl

import (
	"context"
	"testing"
	"time"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

var taskPhase = core.TaskExecution_ABORTED.String()

var taskCreatedAt = time.Date(2018, time.January, 1, 00, 00, 00, 00, time.UTC)
var taskUpdatedAt = time.Date(2018, time.January, 1, 02, 00, 00, 00, time.UTC)
var taskStartedAt = time.Date(2018, time.January, 1, 01, 00, 00, 00, time.UTC)

var retryAttemptValue = uint32(4)

var testTaskExecution = models.TaskExecution{
	TaskExecutionKey: models.TaskExecutionKey{
		TaskKey: models.TaskKey{
			Project: "task project",
			Domain:  "task domain",
			Name:    "task name",
			Version: "task version",
		},
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "2",
			ExecutionKey: models.ExecutionKey{
				Project: "exec project",
				Domain:  "exec domain",
				Name:    "exec name",
			},
		},
		RetryAttempt: &retryAttemptValue,
	},
	Phase:                  taskPhase,
	InputURI:               "testInput.pb",
	StartedAt:              &taskStartedAt,
	Duration:               time.Hour,
	Closure:                []byte("Test"),
	TaskExecutionCreatedAt: &taskCreatedAt,
	TaskExecutionUpdatedAt: &taskUpdatedAt,
}

func getMockTaskExecutionResponseFromDb(expected models.TaskExecution) map[string]interface{} {
	taskExecution := make(map[string]interface{})
	taskExecution["project"] = expected.TaskKey.Project
	taskExecution["domain"] = expected.TaskKey.Domain
	taskExecution["name"] = expected.TaskKey.Name
	taskExecution["version"] = expected.TaskKey.Version
	taskExecution["node_id"] = expected.NodeExecutionKey.NodeID
	taskExecution["execution_project"] = expected.NodeExecutionKey.ExecutionKey.Project
	taskExecution["execution_domain"] = expected.NodeExecutionKey.ExecutionKey.Domain
	taskExecution["execution_name"] = expected.NodeExecutionKey.ExecutionKey.Name
	taskExecution["retry_attempt"] = expected.TaskExecutionKey.RetryAttempt

	taskExecution["phase"] = expected.Phase
	taskExecution["input_uri"] = expected.InputURI
	taskExecution["started_at"] = expected.StartedAt
	taskExecution["duration"] = expected.Duration
	taskExecution["closure"] = expected.Closure
	taskExecution["task_execution_created_at"] = expected.TaskExecutionCreatedAt
	taskExecution["task_execution_updated_at"] = expected.TaskExecutionUpdatedAt
	return taskExecution
}

func TestCreateTaskExecution(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := taskExecutionRepo.Create(context.Background(), testTaskExecution)
	assert.NoError(t, err)
}

func TestUpdateTaskExecution(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	taskExecutionQuery := GlobalMock.NewMock()
	taskExecutionQuery.WithQuery(`UPDATE "task_executions" SET "id"=$1,"created_at"=$2,"updated_at"=$3,"deleted_at"=$4,"phase"=$5,"phase_version"=$6,"input_uri"=$7,"closure"=$8,"started_at"=$9,"task_execution_created_at"=$10,"task_execution_updated_at"=$11,"duration"=$12 WHERE "project" = $13 AND "domain" = $14 AND "name" = $15 AND "version" = $16 AND "execution_project" = $17 AND "execution_domain" = $18 AND "execution_name" = $19 AND "node_id" = $20 AND "retry_attempt" = $21`)
	err := taskExecutionRepo.Update(context.Background(), testTaskExecution)
	assert.NoError(t, err)
	assert.True(t, taskExecutionQuery.Triggered)
}

func TestGetTaskExecution(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	taskExecutions := make([]map[string]interface{}, 0)
	taskExecution := getMockTaskExecutionResponseFromDb(testTaskExecution)
	taskExecutions = append(taskExecutions, taskExecution)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "task_executions" WHERE "task_executions"."project" = $1 AND "task_executions"."domain" = $2 AND "task_executions"."name" = $3 AND "task_executions"."version" = $4 AND "task_executions"."execution_project" = $5 AND "task_executions"."execution_domain" = $6 AND "task_executions"."execution_name" = $7 AND "task_executions"."node_id" = $8 AND "task_executions"."retry_attempt" = $9 LIMIT 1`).
		WithReply(taskExecutions)

	output, err := taskExecutionRepo.Get(context.Background(), interfaces.GetTaskExecutionInput{
		TaskExecutionID: core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "task-id",
				Version:      "task-version",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "node-id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, testTaskExecution, output)
}

func TestListTaskExecutionForExecution(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	taskExecutions := make([]map[string]interface{}, 0)
	taskExecution := getMockTaskExecutionResponseFromDb(testTaskExecution)
	taskExecutions = append(taskExecutions, taskExecution)
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	GlobalMock.NewMock().WithQuery(`SELECT "task_executions"."id","task_executions"."created_at","task_executions"."updated_at","task_executions"."deleted_at","task_executions"."project","task_executions"."domain","task_executions"."name","task_executions"."version","task_executions"."execution_project","task_executions"."execution_domain","task_executions"."execution_name","task_executions"."node_id","task_executions"."retry_attempt","task_executions"."phase","task_executions"."phase_version","task_executions"."input_uri","task_executions"."closure","task_executions"."started_at","task_executions"."task_execution_created_at","task_executions"."task_execution_updated_at","task_executions"."duration" FROM "task_executions" LEFT JOIN tasks ON task_executions.project = tasks.project AND task_executions.domain = tasks.domain AND task_executions.name = tasks.name AND task_executions.version = tasks.version INNER JOIN node_executions ON task_executions.node_id = node_executions.node_id AND task_executions.execution_project = node_executions.execution_project AND task_executions.execution_domain = node_executions.execution_domain AND task_executions.execution_name = node_executions.execution_name INNER JOIN executions ON node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE executions.execution_project = $1 AND executions.execution_domain = $2 AND executions.execution_name = $3 LIMIT 20`).WithReply(taskExecutions)

	collection, err := taskExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Execution, "project", "project_name"),
			getEqualityFilter(common.Execution, "domain", "domain_name"),
			getEqualityFilter(common.Execution, "name", "execution_name"),
		},
		Limit: 20,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.TaskExecutions)
	assert.Len(t, collection.TaskExecutions, 1)

	for _, taskExecution := range collection.TaskExecutions {
		assert.Equal(t, testTaskExecution.TaskExecutionKey, taskExecution.TaskExecutionKey)
		assert.Equal(t, taskPhase, taskExecution.Phase)
		assert.Equal(t, []byte("Test"), taskExecution.Closure)
		assert.Equal(t, "testInput.pb", taskExecution.InputURI)
		assert.Equal(t, taskStartedAt, *taskExecution.StartedAt)
		assert.Equal(t, time.Hour, taskExecution.Duration)
	}
}

func TestListTaskExecutionsForTaskExecution(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	taskExecutions := make([]map[string]interface{}, 0)
	taskExecution := getMockTaskExecutionResponseFromDb(testTaskExecution)
	taskExecutions = append(taskExecutions, taskExecution)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(`SELECT "task_executions"."id","task_executions"."created_at","task_executions"."updated_at","task_executions"."deleted_at","task_executions"."project","task_executions"."domain","task_executions"."name","task_executions"."version","task_executions"."execution_project","task_executions"."execution_domain","task_executions"."execution_name","task_executions"."node_id","task_executions"."retry_attempt","task_executions"."phase","task_executions"."phase_version","task_executions"."input_uri","task_executions"."closure","task_executions"."started_at","task_executions"."task_execution_created_at","task_executions"."task_execution_updated_at","task_executions"."duration" FROM "task_executions" LEFT JOIN tasks ON task_executions.project = tasks.project AND task_executions.domain = tasks.domain AND task_executions.name = tasks.name AND task_executions.version = tasks.version INNER JOIN node_executions ON task_executions.node_id = node_executions.node_id AND task_executions.execution_project = node_executions.execution_project AND task_executions.execution_domain = node_executions.execution_domain AND task_executions.execution_name = node_executions.execution_name INNER JOIN executions ON node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE tasks.project = $1 AND tasks.domain = $2 AND tasks.name = $3 AND tasks.version = $4 AND node_executions.phase = $5 AND executions.execution_project = $6 AND executions.execution_domain = $7 AND executions.execution_name = $8 LIMIT 20`).WithReply(taskExecutions)

	collection, err := taskExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", "project_tn"),
			getEqualityFilter(common.Task, "domain", "domain_t"),
			getEqualityFilter(common.Task, "name", "domain_t"),
			getEqualityFilter(common.Task, "version", "version_t"),

			getEqualityFilter(common.NodeExecution, "phase", nodePhase),
			getEqualityFilter(common.Execution, "project", "project_name"),
			getEqualityFilter(common.Execution, "domain", "domain_name"),
			getEqualityFilter(common.Execution, "name", "execution_name"),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.TaskExecutions)
	assert.Len(t, collection.TaskExecutions, 1)

	for _, taskExecution := range collection.TaskExecutions {
		assert.Equal(t, testTaskExecution.TaskExecutionKey, taskExecution.TaskExecutionKey)
		assert.Equal(t, &retryAttemptValue, taskExecution.RetryAttempt)
		assert.Equal(t, taskPhase, taskExecution.Phase)
		assert.Equal(t, []byte("Test"), taskExecution.Closure)
		assert.Equal(t, "testInput.pb", taskExecution.InputURI)
		assert.Equal(t, taskStartedAt, *taskExecution.StartedAt)
		assert.Equal(t, time.Hour, taskExecution.Duration)
	}
}

func TestCountTaskExecutions(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT count(*) FROM "task_executions"`).WithReply([]map[string]interface{}{{"rows": 2}})

	count, err := taskExecutionRepo.Count(context.Background(), interfaces.CountResourceInput{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCountTaskExecutions_Filters(t *testing.T) {
	taskExecutionRepo := NewTaskExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT count(*) FROM "task_executions" LEFT JOIN tasks ON task_executions.project = tasks.project AND task_executions.domain = tasks.domain AND task_executions.name = tasks.name AND task_executions.version = tasks.version INNER JOIN node_executions ON task_executions.node_id = node_executions.node_id AND task_executions.execution_project = node_executions.execution_project AND task_executions.execution_domain = node_executions.execution_domain AND task_executions.execution_name = node_executions.execution_name INNER JOIN executions ON node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE task_executions.phase = $1 AND "task_execution_updated_at" IS NULL`).WithReply([]map[string]interface{}{{"rows": 3}})

	count, err := taskExecutionRepo.Count(context.Background(), interfaces.CountResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.TaskExecution, "phase", core.TaskExecution_FAILED.String()),
		},
		MapFilters: []common.MapFilter{
			common.NewMapFilter(map[string]interface{}{
				"task_execution_updated_at": nil,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
