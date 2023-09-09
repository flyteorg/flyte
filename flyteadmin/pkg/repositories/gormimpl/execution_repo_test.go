package gormimpl

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
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

func TestUpdateExecution(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	updated := false

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`UPDATE "executions" SET "updated_at"=$1,"execution_project"=$2,` +
		`"execution_domain"=$3,"execution_name"=$4,"launch_plan_id"=$5,"workflow_id"=$6,"phase"=$7,"closure"=$8,` +
		`"spec"=$9,"started_at"=$10,"execution_created_at"=$11,"execution_updated_at"=$12,"duration"=$13 WHERE "` +
		`execution_project" = $14 AND "execution_domain" = $15 AND "execution_name" = $16`).WithCallback(
		func(s string, values []driver.NamedValue) {
			updated = true
		},
	)

	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	//	`WHERE "executions"."deleted_at" IS NULL`)
	err := executionRepo.Update(context.Background(),
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
	assert.True(t, updated)
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
	execution["launch_entity"] = expected.LaunchEntity
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
		LaunchEntity:       "task",
	}

	executions := make([]map[string]interface{}, 0)
	execution := getMockExecutionResponseFromDb(expectedExecution)
	executions = append(executions, execution)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "executions" WHERE "executions"."execution_project" = $1 AND "executions"."execution_domain" = $2 AND "executions"."execution_name" = $3 LIMIT 1`).WithReply(executions)

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
	GlobalMock.NewMock().WithQuery(`SELECT * FROM "executions" WHERE executions.execution_project = $1 AND executions.execution_domain = $2 AND executions.execution_name = $3 AND executions.workflow_id = $4 LIMIT 20`).WithReply(executions[0:1])

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
	mockQuery := GlobalMock.NewMock().WithQuery(`execution_name asc`)
	mockQuery.WithReply(executions)

	sortParameter, err := common.NewSortParameter(&admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "execution_name",
	}, models.ExecutionColumns)
	require.NoError(t, err)

	_, err = executionRepo.List(context.Background(), interfaces.ListResourceInput{
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

func TestListExecutions_WithTags(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	executions := make([]map[string]interface{}, 0)
	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that include ordering by name
	mockQuery := GlobalMock.NewMock().WithQuery(`execution_name asc`)
	mockQuery.WithReply(executions)

	sortParameter, err := common.NewSortParameter(&admin.Sort{
		Direction: admin.Sort_ASCENDING,
		Key:       "execution_name",
	}, models.ExecutionColumns)
	require.NoError(t, err)

	vals := []string{"tag1", "tag2"}
	tagFilter, err := common.NewRepeatedValueFilter(common.ExecutionAdminTag, common.ValueIn, "admin_tag_name", vals)
	assert.NoError(t, err)

	_, err = executionRepo.List(context.Background(), interfaces.ListResourceInput{
		SortParameter: sortParameter,
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Task, "project", project),
			getEqualityFilter(common.Task, "domain", domain),
			getEqualityFilter(common.Task, "name", name),
			tagFilter,
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
		LaunchEntity: "launch_plan",
		Tags:         []models.AdminTag{{Name: "tag1"}, {Name: "tag2"}},
	})
	executions = append(executions, execution)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true
	// Only match on queries that append expected filters
	GlobalMock.NewMock().WithQuery(`SELECT "executions"."id","executions"."created_at","executions"."updated_at","executions"."deleted_at","executions"."execution_project","executions"."execution_domain","executions"."execution_name","executions"."launch_plan_id","executions"."workflow_id","executions"."task_id","executions"."phase","executions"."closure","executions"."spec","executions"."started_at","executions"."execution_created_at","executions"."execution_updated_at","executions"."duration","executions"."abort_cause","executions"."mode","executions"."source_execution_id","executions"."parent_node_execution_id","executions"."cluster","executions"."inputs_uri","executions"."user_inputs_uri","executions"."error_kind","executions"."error_code","executions"."user","executions"."state","executions"."launch_entity" FROM "executions" INNER JOIN workflows ON executions.workflow_id = workflows.id INNER JOIN tasks ON executions.task_id = tasks.id WHERE executions.execution_project = $1 AND executions.execution_domain = $2 AND executions.execution_name = $3 AND workflows.name = $4 AND tasks.name = $5 AND execution_admin_tags.execution_tag_name in ($6,$7) LIMIT 20`).WithReply(executions)
	vals := []string{"tag1", "tag2"}
	tagFilter, err := common.NewRepeatedValueFilter(common.ExecutionAdminTag, common.ValueIn, "execution_tag_name", vals)
	assert.NoError(t, err)
	collection, err := executionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Execution, "project", project),
			getEqualityFilter(common.Execution, "domain", domain),
			getEqualityFilter(common.Execution, "name", "1"),
			getEqualityFilter(common.Workflow, "name", "workflow_name"),
			getEqualityFilter(common.Task, "name", "task_name"),
			tagFilter,
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
		assert.Equal(t, "launch_plan", execution.LaunchEntity)
	}
}

func TestCountExecutions(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT count(*) FROM "executions"`).WithReply([]map[string]interface{}{{"rows": 2}})

	count, err := executionRepo.Count(context.Background(), interfaces.CountResourceInput{})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestCountExecutions_Filters(t *testing.T) {
	executionRepo := NewExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT count(*) FROM "executions" INNER JOIN workflows ON executions.workflow_id = workflows.id INNER JOIN tasks ON executions.task_id = tasks.id WHERE executions.phase = $1 AND "error_code" IS NULL`,
	).WithReply([]map[string]interface{}{{"rows": 3}})

	count, err := executionRepo.Count(context.Background(), interfaces.CountResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.Execution, "phase", core.WorkflowExecution_FAILED.String()),
		},
		MapFilters: []common.MapFilter{
			common.NewMapFilter(map[string]interface{}{
				"error_code": nil,
			}),
		},
		JoinTableEntities: map[common.Entity]bool{
			common.Workflow: true,
			common.Task:     true,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
