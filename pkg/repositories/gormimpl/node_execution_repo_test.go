package gormimpl

import (
	"context"
	"testing"
	"time"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/stretchr/testify/assert"
)

var nodePhase = core.NodeExecution_RUNNING.String()

var nodeStartedAt = time.Date(2018, time.February, 17, 00, 00, 00, 00, time.UTC)

var nodeCreatedAt = time.Date(2018, time.February, 17, 00, 00, 00, 00, time.UTC).UTC()
var nodePlanUpdatedAt = time.Date(2018, time.February, 17, 00, 01, 00, 00, time.UTC).UTC()

func TestCreateNodeExecution(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()

	nodeExecutionQuery := GlobalMock.NewMock()
	nodeExecutionQuery.WithQuery(`INSERT INTO "node_executions" ("created_at","updated_at","deleted_at","execution_project","execution_domain","execution_name","node_id","phase","input_uri","closure","started_at","node_execution_created_at","node_execution_updated_at","duration","node_execution_metadata","parent_id","parent_task_execution_id","error_kind","error_code","cache_status","dynamic_workflow_remote_closure_reference") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`)

	parentID := uint(10)
	nodeExecution := models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "1",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
		},
		Phase:                  nodePhase,
		Closure:                []byte("closure"),
		NodeExecutionMetadata:  []byte("closure"),
		InputURI:               "input uri",
		StartedAt:              &nodeStartedAt,
		Duration:               time.Hour,
		NodeExecutionCreatedAt: &nodeCreatedAt,
		NodeExecutionUpdatedAt: &nodeCreatedAt,
		ParentID:               &parentID,
	}
	err := nodeExecutionRepo.Create(context.Background(), &nodeExecution)
	assert.NoError(t, err)
	assert.True(t, nodeExecutionQuery.Triggered)
}

func TestUpdateNodeExecution(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append the name filter
	nodeExecutionQuery := GlobalMock.NewMock()
	nodeExecutionQuery.WithQuery(`UPDATE "node_executions" SET "id"=$1,"updated_at"=$2,"execution_project"=$3,"execution_domain"=$4,"execution_name"=$5,"node_id"=$6,"phase"=$7,"input_uri"=$8,"closure"=$9,"started_at"=$10,"node_execution_created_at"=$11,"node_execution_updated_at"=$12,"duration"=$13 WHERE "execution_project" = $14 AND "execution_domain" = $15 AND "execution_name" = $16 AND "node_id" = $17`)
	err := nodeExecutionRepo.Update(context.Background(),
		&models.NodeExecution{
			BaseModel: models.BaseModel{ID: 1},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "1",
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "1",
				},
			},
			Phase:                  nodePhase,
			Closure:                []byte("closure"),
			InputURI:               "input uri",
			StartedAt:              &nodeStartedAt,
			Duration:               time.Hour,
			NodeExecutionCreatedAt: &nodeCreatedAt,
			NodeExecutionUpdatedAt: &nodePlanUpdatedAt,
		})
	assert.NoError(t, err)
	assert.True(t, nodeExecutionQuery.Triggered)
}

func getMockNodeExecutionResponseFromDb(expected models.NodeExecution) map[string]interface{} {
	nodeExecution := make(map[string]interface{})
	nodeExecution["execution_project"] = expected.ExecutionKey.Project
	nodeExecution["execution_domain"] = expected.ExecutionKey.Domain
	nodeExecution["execution_name"] = expected.ExecutionKey.Name
	nodeExecution["node_id"] = expected.NodeExecutionKey.NodeID
	nodeExecution["phase"] = expected.Phase
	nodeExecution["closure"] = expected.Closure
	nodeExecution["input_uri"] = expected.InputURI
	nodeExecution["started_at"] = expected.StartedAt
	nodeExecution["duration"] = expected.Duration
	nodeExecution["node_execution_created_at"] = expected.NodeExecutionCreatedAt
	nodeExecution["node_execution_updated_at"] = expected.NodeExecutionUpdatedAt
	nodeExecution["parent_id"] = expected.ParentID
	if expected.NodeExecutionMetadata != nil {
		nodeExecution["node_execution_metadata"] = expected.NodeExecutionMetadata
	}
	return nodeExecution
}

func TestGetNodeExecution(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	parentID := uint(10)
	expectedNodeExecution := models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "1",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
		},
		Phase:                  nodePhase,
		Closure:                []byte("closure"),
		InputURI:               "input uri",
		StartedAt:              &nodeStartedAt,
		Duration:               time.Hour,
		NodeExecutionCreatedAt: &nodeCreatedAt,
		NodeExecutionUpdatedAt: &nodePlanUpdatedAt,
		NodeExecutionMetadata:  []byte("NodeExecutionMetadata"),
		ParentID:               &parentID,
	}

	nodeExecutions := make([]map[string]interface{}, 0)
	nodeExecution := getMockNodeExecutionResponseFromDb(expectedNodeExecution)
	nodeExecutions = append(nodeExecutions, nodeExecution)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT * FROM "node_executions" WHERE "node_executions"."execution_project" = $1 AND "node_executions"."execution_domain" = $2 AND "node_executions"."execution_name" = $3 AND "node_executions"."node_id" = $4 LIMIT 1`).WithReply(nodeExecutions)
	output, err := nodeExecutionRepo.Get(context.Background(), interfaces.NodeExecutionResource{
		NodeExecutionIdentifier: core.NodeExecutionIdentifier{
			NodeId: "1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "execution_project",
				Domain:  "execution_domain",
				Name:    "execution_name",
			},
		},
	})
	assert.NoError(t, err)
	assert.EqualValues(t, expectedNodeExecution, output)
}

func TestListNodeExecutions(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	nodeExecutions := make([]map[string]interface{}, 0)
	executionIDs := []string{"100", "200"}
	for _, executionID := range executionIDs {
		nodeExecution := getMockNodeExecutionResponseFromDb(models.NodeExecution{
			NodeExecutionKey: models.NodeExecutionKey{
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    executionID,
				},
			},
			Phase:                  nodePhase,
			Closure:                []byte("closure"),
			InputURI:               "input uri",
			StartedAt:              &nodeStartedAt,
			Duration:               time.Hour,
			NodeExecutionCreatedAt: &nodeCreatedAt,
			NodeExecutionUpdatedAt: &nodePlanUpdatedAt,
		})
		nodeExecutions = append(nodeExecutions, nodeExecution)
	}

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(`SELECT "node_executions"."id","node_executions"."created_at","node_executions"."updated_at","node_executions"."deleted_at","node_executions"."execution_project","node_executions"."execution_domain","node_executions"."execution_name","node_executions"."node_id","node_executions"."phase","node_executions"."input_uri","node_executions"."closure","node_executions"."started_at","node_executions"."node_execution_created_at","node_executions"."node_execution_updated_at","node_executions"."duration","node_executions"."node_execution_metadata","node_executions"."parent_id","node_executions"."parent_task_execution_id","node_executions"."error_kind","node_executions"."error_code","node_executions"."cache_status","node_executions"."dynamic_workflow_remote_closure_reference" FROM "node_executions" INNER JOIN executions ON node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE node_executions.phase = $1 LIMIT 20`).
		WithReply(nodeExecutions)

	collection, err := nodeExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.NodeExecution, "phase", nodePhase),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.NodeExecutions)
	assert.Len(t, collection.NodeExecutions, 2)
	for _, nodeExecution := range collection.NodeExecutions {
		assert.Equal(t, "project", nodeExecution.ExecutionKey.Project)
		assert.Equal(t, "domain", nodeExecution.ExecutionKey.Domain)
		assert.Contains(t, executionIDs, nodeExecution.ExecutionKey.Name)
		assert.Equal(t, nodePhase, nodeExecution.Phase)
		assert.Equal(t, []byte("closure"), nodeExecution.Closure)
		assert.Equal(t, "input uri", nodeExecution.InputURI)
		assert.Equal(t, nodeStartedAt, *nodeExecution.StartedAt)
		assert.Equal(t, time.Hour, nodeExecution.Duration)
	}
}

func TestListNodeExecutions_Order(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	nodeExecutions := make([]map[string]interface{}, 0)

	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that include ordering by project
	mockQuery := GlobalMock.NewMock()
	mockQuery.WithQuery(`project desc`)
	mockQuery.WithReply(nodeExecutions)

	sortParameter, _ := common.NewSortParameter(admin.Sort{
		Direction: admin.Sort_DESCENDING,
		Key:       "project",
	})
	_, err := nodeExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		SortParameter: sortParameter,
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.NodeExecution, "phase", nodePhase),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.True(t, mockQuery.Triggered)
}

func TestListNodeExecutions_MissingParameters(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := nodeExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.NodeExecution, "execution_id", "1234"),
		},
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: limit")

	_, err = nodeExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit: 20,
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: filters")
}

func TestListNodeExecutionsForExecution(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	nodeExecutions := make([]map[string]interface{}, 0)
	nodeExecution := getMockNodeExecutionResponseFromDb(models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
		},
		Phase:                 nodePhase,
		Closure:               []byte("closure"),
		InputURI:              "input uri",
		StartedAt:             &nodeStartedAt,
		Duration:              time.Hour,
		NodeExecutionMetadata: []byte("NodeExecutionMetadata"),
	})
	nodeExecutions = append(nodeExecutions, nodeExecution)

	GlobalMock := mocket.Catcher.Reset()
	query := `SELECT "node_executions"."id","node_executions"."created_at","node_executions"."updated_at","node_executions"."deleted_at","node_executions"."execution_project","node_executions"."execution_domain","node_executions"."execution_name","node_executions"."node_id","node_executions"."phase","node_executions"."input_uri","node_executions"."closure","node_executions"."started_at","node_executions"."node_execution_created_at","node_executions"."node_execution_updated_at","node_executions"."duration","node_executions"."node_execution_metadata","node_executions"."parent_id","node_executions"."parent_task_execution_id","node_executions"."error_kind","node_executions"."error_code","node_executions"."cache_status","node_executions"."dynamic_workflow_remote_closure_reference" FROM "node_executions" INNER JOIN executions ON node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE node_executions.phase = $1 AND executions.execution_name = $2 LIMIT 20`
	GlobalMock.NewMock().WithQuery(query).WithReply(nodeExecutions)

	collection, err := nodeExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.NodeExecution, "phase", nodePhase),
			getEqualityFilter(common.Execution, "name", "execution_name"),
		},

		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.NodeExecutions)
	assert.Len(t, collection.NodeExecutions, 1)
	for _, nodeExecution := range collection.NodeExecutions {
		assert.Equal(t, "project", nodeExecution.ExecutionKey.Project)
		assert.Equal(t, "domain", nodeExecution.ExecutionKey.Domain)
		assert.Equal(t, "1", nodeExecution.ExecutionKey.Name)
		assert.Equal(t, nodePhase, nodeExecution.Phase)
		assert.Equal(t, []byte("closure"), nodeExecution.Closure)
		assert.Equal(t, "input uri", nodeExecution.InputURI)
		assert.Equal(t, nodeStartedAt, *nodeExecution.StartedAt)
		assert.Equal(t, time.Hour, nodeExecution.Duration)
		assert.Equal(t, []byte("NodeExecutionMetadata"), nodeExecution.NodeExecutionMetadata)
		assert.Empty(t, nodeExecution.ChildNodeExecutions)
		assert.Empty(t, nodeExecution.ParentID)
	}
}

func getMockNodeExecutionEventResponseFromDb(expected models.NodeExecutionEvent) map[string]interface{} {
	nodeExecutionEvent := make(map[string]interface{})
	nodeExecutionEvent["execution_project"] = expected.ExecutionKey.Project
	nodeExecutionEvent["execution_domain"] = expected.ExecutionKey.Domain
	nodeExecutionEvent["execution_name"] = expected.ExecutionKey.Name
	nodeExecutionEvent["node_id"] = expected.NodeExecutionKey.NodeID
	nodeExecutionEvent["request_id"] = expected.RequestID
	nodeExecutionEvent["phase"] = expected.Phase
	nodeExecutionEvent["occurred_at"] = expected.OccurredAt
	return nodeExecutionEvent
}

func TestListNodeExecutionEventsForNodeExecutionAndExecution(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())

	nodeExecutions := make([]map[string]interface{}, 0)
	nodeExecution := getMockNodeExecutionEventResponseFromDb(models.NodeExecutionEvent{
		NodeExecutionKey: models.NodeExecutionKey{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
		},
		RequestID:  "1",
		Phase:      nodePhase,
		OccurredAt: nodeStartedAt,
	})
	nodeExecutions = append(nodeExecutions, nodeExecution)

	GlobalMock := mocket.Catcher.Reset()
	query := `SELECT "node_execution_events"."id","node_execution_events"."created_at","node_execution_events"."updated_at","node_execution_events"."deleted_at","node_execution_events"."execution_project","node_execution_events"."execution_domain","node_execution_events"."execution_name","node_execution_events"."node_id","node_execution_events"."request_id","node_execution_events"."occurred_at","node_execution_events"."phase" FROM "node_execution_events" INNER JOIN node_executions ON node_event_executions.node_execution_id = node_executions.id INNER JOIN executions ON node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE node_executions.execution_id = $1 AND node_executions.node_id = $2 AND node_execution_events.request_id = $3 AND executions.execution_project = $4 AND executions.execution_domain = $5 AND executions.execution_name = $6 LIMIT 20`
	GlobalMock.NewMock().WithQuery(query).WithReply(nodeExecutions)

	collection, err := nodeExecutionRepo.ListEvents(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.NodeExecution, "execution_id", uint(1)),
			getEqualityFilter(common.NodeExecution, "node_id", uint(2)),
			getEqualityFilter(common.NodeExecutionEvent, "request_id", "1"),
			getEqualityFilter(common.Execution, "project", project),
			getEqualityFilter(common.Execution, "domain", domain),
			getEqualityFilter(common.Execution, "name", name),
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, collection)
	assert.NotEmpty(t, collection.NodeExecutionEvents)
	assert.Len(t, collection.NodeExecutionEvents, 1)
	for _, event := range collection.NodeExecutionEvents {
		assert.Equal(t, "1", event.RequestID)
		assert.Equal(t, models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		}, event.ExecutionKey)
		assert.Equal(t, nodePhase, event.Phase)
		assert.Equal(t, nodeStartedAt, event.OccurredAt)
	}
}

func TestListNodeExecutionEvents_MissingParameters(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	_, err := nodeExecutionRepo.ListEvents(context.Background(), interfaces.ListResourceInput{
		InlineFilters: []common.InlineFilter{
			getEqualityFilter(common.NodeExecution, "node_id", "1234"),
		},
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: limit")

	_, err = nodeExecutionRepo.List(context.Background(), interfaces.ListResourceInput{
		Limit: 20,
	})
	assert.EqualError(t, err, "missing and/or invalid parameters: filters")
}

func TestNodeExecutionExists(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	id := uint(10)
	expectedNodeExecution := models.NodeExecution{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "1",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
		},
		BaseModel: models.BaseModel{
			ID: id,
		},
		Phase:   nodePhase,
		Closure: []byte("closure"),
	}

	nodeExecutions := make([]map[string]interface{}, 0)
	nodeExecution := getMockNodeExecutionResponseFromDb(expectedNodeExecution)
	nodeExecutions = append(nodeExecutions, nodeExecution)

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.NewMock().WithQuery(
		`SELECT "id" FROM "node_executions" WHERE "node_executions"."execution_project" = $1 AND "node_executions"."execution_domain" = $2 AND "node_executions"."execution_name" = $3 AND "node_executions"."node_id" = $4 LIMIT 1`).WithReply(nodeExecutions)
	exists, err := nodeExecutionRepo.Exists(context.Background(), interfaces.NodeExecutionResource{
		NodeExecutionIdentifier: core.NodeExecutionIdentifier{
			NodeId: "1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "execution_project",
				Domain:  "execution_domain",
				Name:    "execution_name",
			},
		},
	})
	assert.NoError(t, err)
	assert.True(t, exists)
}
