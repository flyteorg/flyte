package gormimpl

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/stretchr/testify/assert"
)

var createdAt = time.Date(2018, time.February, 17, 00, 00, 00, 00, time.UTC).UTC()
var executionStartedAt = time.Date(2018, time.February, 17, 00, 01, 00, 00, time.UTC).UTC()
var executionUpdatedAt = time.Date(2018, time.February, 17, 00, 01, 00, 00, time.UTC).UTC()

func TestCreateExecution(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := executionRepo.Create(context.Background(), models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		},
		LaunchPlanID:       uint(2),
		Phase:              core.WorkflowExecution_SUCCEEDED.String(),
		Closure:            []byte{1, 2},
		Spec:               []byte{3, 4},
		StartedAt:          &executionStartedAt,
		ExecutionCreatedAt: &createdAt,
	})
	assert.NoError(t, err)
}

func TestUpdate(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	executionEventQuery := GlobalMock.NewMock()
	executionEventQuery.WithQuery(`INSERT INTO "execution_events" ("created_at","updated_at","deleted_at",` +
		`"execution_project","execution_domain","execution_name","request_id","occurred_at","phase") VALUES ` +
		`(?,?,?,?,?,?,?,?,?)`)
	executionQuery := GlobalMock.NewMock()
	executionQuery.WithQuery(`UPDATE "executions" SET "closure" = ?, "duration" = ?, "execution_created_at" = ?, ` +
		`"execution_domain" = ?, "execution_name" = ?, "execution_project" = ?, "execution_updated_at" = ?, ` +
		`"launch_plan_id" = ?, "mode" = ?, "phase" = ?, "spec" = ?, "started_at" = ?, "updated_at" = ?, ` +
		`"workflow_id" = ?  WHERE "executions"."deleted_at" IS NULL`)
	err := executionRepo.Update(context.Background(),
		models.ExecutionEvent{
			RequestID: "request id 1",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
			OccurredAt: time.Now(),
			Phase:      core.WorkflowExecution_SUCCEEDED.String(),
		},
		models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
			LaunchPlanID:       uint(2),
			WorkflowID:         uint(3),
			Phase:              core.WorkflowExecution_SUCCEEDED.String(),
			Closure:            []byte{1, 2},
			Spec:               []byte{3, 4},
			StartedAt:          &executionStartedAt,
			ExecutionCreatedAt: &createdAt,
			ExecutionUpdatedAt: &executionUpdatedAt,
			Duration:           time.Hour,
			Mode:               1,
		})
	assert.NoError(t, err)
	assert.True(t, executionEventQuery.Triggered)
	assert.True(t, executionQuery.Triggered)
}

func TestUpdateExecution(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	executionQuery := GlobalMock.NewMock()
	executionQuery.WithQuery(`UPDATE "executions" SET "closure" = ?, "duration" = ?, "execution_created_at" = ?, ` +
		`"execution_domain" = ?, "execution_name" = ?, "execution_project" = ?, "execution_updated_at" = ?, ` +
		`"launch_plan_id" = ?, "phase" = ?, "spec" = ?, "started_at" = ?, "updated_at" = ?, "workflow_id" = ?  ` +
		`WHERE "executions"."deleted_at" IS NULL`)
	err := executionRepo.UpdateExecution(context.Background(),
		models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
			LaunchPlanID:       uint(2),
			WorkflowID:         uint(3),
			Phase:              core.WorkflowExecution_SUCCEEDED.String(),
			Closure:            []byte{1, 2},
			Spec:               []byte{3, 4},
			StartedAt:          &executionStartedAt,
			ExecutionCreatedAt: &createdAt,
			ExecutionUpdatedAt: &executionUpdatedAt,
			Duration:           time.Hour,
		})
	assert.NoError(t, err)
	assert.True(t, executionQuery.Triggered)
}

func getMockExecutionResponseFromDb(expected models.Execution) map[string]interface{} {
	execution := make(map[string]interface{})
	execution["id"] = expected.ID
	execution["execution_project"] = expected.Project
	execution["execution_domain"] = expected.Domain
	execution["execution_name"] = expected.Name
	execution["launch_plan_id"] = expected.LaunchPlanID
	execution["workflow_id"] = expected.WorkflowID
	execution["phase"] = expected.Phase
	execution["closure"] = expected.Closure
	execution["spec"] = expected.Spec
	execution["started_at"] = expected.StartedAt
	execution["execution_created_at"] = expected.ExecutionCreatedAt
	execution["execution_updated_at"] = expected.ExecutionUpdatedAt
	execution["duration"] = expected.Duration
	execution["mode"] = expected.Mode
	return execution
}

func TestGetExecution(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	expectedExecution := models.Execution{
		BaseModel: models.BaseModel{
			ID: uint(20),
		},
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		},
		LaunchPlanID:       uint(2),
		Phase:              core.WorkflowExecution_SUCCEEDED.String(),
		Closure:            []byte{1, 2},
		WorkflowID:         uint(3),
		Spec:               []byte{3, 4},
		StartedAt:          &executionStartedAt,
		ExecutionCreatedAt: &createdAt,
		ExecutionUpdatedAt: &executionUpdatedAt,
	}

	executions := make([]map[string]interface{}, 0)
	execution := getMockExecutionResponseFromDb(expectedExecution)
	executions = append(executions, execution)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "executions"  WHERE "executions"."deleted_at" IS NULL AND ` +
		`(("executions"."execution_project" = project) AND ("executions"."execution_domain" = domain) AND ` +
		`("executions"."execution_name" = 1)) LIMIT 1`).WithReply(executions)
	output, err := executionRepo.Get(context.Background(), interfaces.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "1",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, expectedExecution, output)
}

func TestListExecutions(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	executions := make([]map[string]interface{}, 0)
	names := []string{"ABC", "XYZ"}
	for _, name := range names {
		execution := getMockExecutionResponseFromDb(models.Execution{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    name,
			},
			LaunchPlanID: uint(2),
			WorkflowID:   uint(3),
			Phase:        core.WorkflowExecution_SUCCEEDED.String(),
			Closure:      []byte{1, 2},
			Spec:         []byte{3, 4},
			StartedAt:    &executionStartedAt,
			Duration:     time.Hour,
		})
		executions = append(executions, execution)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithReply(executions)

	collection, err := executionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Executions)
	assert.Len(t, collection.Executions, 2)
	for _, execution := range collection.Executions {
		assert.Equal(t, project, execution.Project)
		assert.Equal(t, domain, execution.Domain)
		assert.Contains(t, names, execution.Name)
		assert.Equal(t, uint(2), execution.LaunchPlanID)
		assert.Equal(t, uint(3), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_SUCCEEDED.String(), execution.Phase)
		assert.Equal(t, []byte{1, 2}, execution.Closure)
		assert.Equal(t, []byte{3, 4}, execution.Spec)
		assert.Equal(t, executionStartedAt, *execution.StartedAt)
		assert.Equal(t, time.Hour, execution.Duration)
	}
}

func TestListExecutions_Filters(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	executions := make([]map[string]interface{}, 0)
	execution := getMockExecutionResponseFromDb(models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		},
		LaunchPlanID: uint(2),
		WorkflowID:   uint(3),
		Phase:        core.WorkflowExecution_SUCCEEDED.String(),
		Closure:      []byte{1, 2},
		Spec:         []byte{3, 4},
		StartedAt:    &executionStartedAt,
		Duration:     time.Hour,
	})
	executions = append(executions, execution)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append the name filter
	GlobalMock.NewMock().WithQuery(`name = 1`).WithReply(executions[0:1])

	collection, err := executionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Execution, "project", project),
			getEqualityFilter(common.Execution, "domain", domain),
			getEqualityFilter(common.Execution, "name", "1"),
			getEqualityFilter(common.Execution, "workflow_id", workflowID),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Executions)
	assert.Len(t, collection.Executions, 1)

	result := collection.Executions[0]
	assert.Equal(t, project, result.Project)
	assert.Equal(t, domain, result.Domain)
	assert.Equal(t, "1", result.Name)
	assert.Equal(t, uint(2), result.LaunchPlanID)
	assert.Equal(t, uint(3), result.WorkflowID)
	assert.Equal(t, core.WorkflowExecution_SUCCEEDED.String(), result.Phase)
	assert.Equal(t, []byte{1, 2}, result.Closure)
	assert.Equal(t, []byte{3, 4}, result.Spec)
	assert.Equal(t, executionStartedAt, *result.StartedAt)
	assert.Equal(t, time.Hour, result.Duration)
}

func TestListExecutions_Order(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	executions := make([]map[string]interface{}, 0)
	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that include ordering by name
	mockQuery := GlobalMock.NewMock().WithQuery(`name asc`)
	mockQuery.WithReply(executions)

	sortParameter, _ := common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "name",
	})
	_, err := executionRepo.List(context.Background(), interfaces.ListResourceInput{
		SortParameter: sortParameter,
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListExecutions_MissingParameters(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := executionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Execution, "project", project),
			getEqualityFilter(common.Execution, "domain", domain),
			getEqualityFilter(common.Execution, "name", name),
			getEqualityFilter(common.Execution, "workflow_id", workflowID),
		},
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: limit")

	_, err = executionRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit: 20,
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: filters")
}

func TestListExecutionsForWorkflow(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	executions := make([]map[string]interface{}, 0)
	execution := getMockExecutionResponseFromDb(models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		},
		LaunchPlanID: uint(2),
		WorkflowID:   uint(3),
		Phase:        core.WorkflowExecution_SUCCEEDED.String(),
		Closure:      []byte{1, 2},
		Spec:         []byte{3, 4},
		StartedAt:    &executionStartedAt,
		Duration:     time.Hour,
	})
	executions = append(executions, execution)

	GlobalMock := mocket.Catcher.Reset()
	query := `SELECT "executions".* FROM "executions" INNER JOIN workflows ON executions.workflow_id = workflows.id ` +
		`INNER JOIN tasks ON executions.task_id = tasks.id WHERE "executions"."deleted_at" IS NULL AND ` +
		`((executions.execution_project = project) AND (executions.execution_domain = domain) AND ` +
		`(executions.execution_name = 1) AND (workflows.name = workflow_name) AND (tasks.name = task_name)) ` +
		`LIMIT 20 OFFSET 0`
	GlobalMock.NewMock().WithQuery(query).WithReply(executions)

	collection, err := executionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Execution, "project", project),
			getEqualityFilter(common.Execution, "domain", domain),
			getEqualityFilter(common.Execution, "name", "1"),
			getEqualityFilter(common.Workflow, "name", "workflow_name"),
			getEqualityFilter(common.Task, "name", "task_name"),
		},
		Limit: 20,
		JoinTableEntities: map[common.Entity]bool{
			common.Workflow: true,
			common.Task:     true,
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.Executions)
	assert.Len(t, collection.Executions, 1)
	for _, execution := range collection.Executions {
		assert.Equal(t, project, execution.Project)
		assert.Equal(t, domain, execution.Domain)
		assert.Equal(t, "1", execution.Name)
		assert.Equal(t, uint(2), execution.LaunchPlanID)
		assert.Equal(t, uint(3), execution.WorkflowID)
		assert.Equal(t, core.WorkflowExecution_SUCCEEDED.String(), execution.Phase)
		assert.Equal(t, []byte{1, 2}, execution.Closure)
		assert.Equal(t, []byte{3, 4}, execution.Spec)
		assert.Equal(t, executionStartedAt, *execution.StartedAt)
		assert.Equal(t, time.Hour, execution.Duration)
	}
}

func TestExecutionExists(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	expectedExecution := models.Execution{
		BaseModel: models.BaseModel{
			ID: uint(20),
		},
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		},
		LaunchPlanID: uint(2),
		Phase:        core.WorkflowExecution_SUCCEEDED.String(),
		Closure:      []byte{1, 2},
		WorkflowID:   uint(3),
		Spec:         []byte{3, 4},
	}

	executions := make([]map[string]interface{}, 0)
	execution := getMockExecutionResponseFromDb(expectedExecution)
	executions = append(executions, execution)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT id FROM "executions"  WHERE "executions"."deleted_at" IS NULL AND ` +
		`(("executions"."execution_project" = project) AND ("executions"."execution_domain" = domain) AND ` +
		`("executions"."execution_name" = 1)) LIMIT 1`).WithReply(executions)
	exists, err := executionRepo.Exists(context.Background(), interfaces.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "1",
	})
	assert.NoError(t, err)
	assert.True(t, exists)
}
