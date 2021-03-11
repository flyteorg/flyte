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
	nodeExecutionQuery.WithQuery(`INSERT INTO "node_executions" ("id","created_at","updated_at","deleted_at",` +
		`"execution_project","execution_domain","execution_name","node_id","phase","input_uri","closure","started_at",` +
		`"node_execution_created_at","node_execution_updated_at","duration","node_execution_metadata","parent_id","error_kind","error_code","cache_status") VALUES ` +
		`(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)

	nodeExecutionEventQuery := GlobalMock.NewMock()
	nodeExecutionEventQuery.WithQuery(`INSERT INTO "node_execution_events" ("created_at","updated_at",` +
		`"deleted_at","execution_project","execution_domain","execution_name","node_id","request_id","occurred_at",` +
		`"phase") VALUES (?,?,?,?,?,?,?,?,?,?)`)

	nodeExecutionEvent := models.NodeExecutionEvent{
		NodeExecutionKey: models.NodeExecutionKey{
			NodeID: "1",
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "1",
			},
		},
		RequestID:  "xxyzz",
		Phase:      nodePhase,
		OccurredAt: nodeStartedAt,
	}
	parentID := uint(10)
	nodeExecution := models.NodeExecution{
		BaseModel: models.BaseModel{
			ID: 2,
		},
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
	err := nodeExecutionRepo.Create(context.Background(), &nodeExecutionEvent, &nodeExecution)
	assert.NoError(t, err)
	assert.True(t, nodeExecutionEventQuery.Triggered)
	assert.True(t, nodeExecutionQuery.Triggered)
}

func TestUpdateNodeExecution(t *testing.T) {
	nodeExecutionRepo := NewNodeExecutionRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	GlobalMock := mocket.Catcher.Reset()
	// Only match on queries that append the name filter
	nodeExecutionEventQuery := GlobalMock.NewMock()
	nodeExecutionEventQuery.WithQuery(`INSERT INTO "node_execution_events" ("created_at","updated_at",` +
		`"deleted_at","execution_project","execution_domain","execution_name","node_id","request_id","occurred_at",` +
		`"phase") VALUES (?,?,?,?,?,?,?,?,?,?)`)
	nodeExecutionQuery := GlobalMock.NewMock()
	nodeExecutionQuery.WithQuery(`UPDATE "node_executions" SET "closure" = ?, "duration" = ?, ` +
		`"execution_domain" = ?, "execution_name" = ?, "execution_project" = ?, "id" = ?, "input_uri" = ?, ` +
		`"node_execution_created_at" = ?, "node_execution_updated_at" = ?, "node_id" = ?, "phase" = ?, ` +
		`"started_at" = ?, "updated_at" = ?  WHERE "node_executions"."deleted_at" IS NULL AND "node_executions".` +
		`"execution_project" = ? AND "node_executions"."execution_domain" = ? AND "node_executions".` +
		`"execution_name" = ? AND "node_executions"."node_id" = ?`)
	err := nodeExecutionRepo.Update(context.Background(),
		&models.NodeExecutionEvent{
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: "1",
				ExecutionKey: models.ExecutionKey{
					Project: "project",
					Domain:  "domain",
					Name:    "1",
				},
			},
			RequestID:  "request id 1",
			OccurredAt: time.Now(),
			Phase:      nodePhase,
		},
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
	assert.True(t, nodeExecutionEventQuery.Triggered)
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
		`SELECT * FROM "node_executions"  WHERE "node_executions"."deleted_at" IS NULL AND ` +
			`(("node_executions"."execution_project" = execution_project) AND ("node_executions"."execution_domain" ` +
			`= execution_domain) AND ("node_executions"."execution_name" = execution_name) AND ("node_executions".` +
			`"node_id" = 1)) ORDER BY "node_executions"."id" ASC LIMIT 1`).WithReply(nodeExecutions)
	output, err := nodeExecutionRepo.Get(context.Background(), interfaces.GetNodeExecutionInput{
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
	GlobalMock.NewMock().WithQuery(`SELECT "node_executions".* FROM "node_executions" INNER JOIN executions ON ` +
		`node_executions.execution_project = executions.execution_project AND node_executions.execution_domain = ` +
		`executions.execution_domain AND node_executions.execution_name = executions.execution_name WHERE ` +
		`"node_executions"."deleted_at" IS NULL AND ((node_executions.phase = RUNNING)) LIMIT 20 OFFSET 0`).
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
	query := `SELECT "node_executions".* FROM "node_executions" INNER JOIN executions ON node_executions.` +
		`execution_project = executions.execution_project AND node_executions.execution_domain = executions.` +
		`execution_domain AND node_executions.execution_name = executions.execution_name WHERE "node_executions".` +
		`"deleted_at" IS NULL AND ((node_executions.phase = RUNNING) AND ` +
		`(executions.execution_name = execution_name)) LIMIT 20 OFFSET 0`
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
	query := `SELECT "node_execution_events".* FROM "node_execution_events" INNER JOIN node_executions ON ` +
		`node_event_executions.node_execution_id = node_executions.id INNER JOIN executions ON node_executions.` +
		`execution_project = executions.execution_project AND node_executions.execution_domain = executions.` +
		`execution_domain AND node_executions.execution_name = executions.execution_name WHERE ` +
		`"node_execution_events"."deleted_at" IS NULL AND ((node_executions.execution_id = 1) AND ` +
		`(node_executions.node_id = 2) AND (node_execution_events.request_id = 1) AND (executions.execution_project = ` +
		`project) AND (executions.execution_domain = domain) AND (executions.execution_name = name)) LIMIT 20 OFFSET 0`
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
