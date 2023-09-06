package gormimpl

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateNodeExecutionEvent(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	nodeExecutionEventQuery := GlobalMock.NewMock()
	nodeExecutionEventQuery.WithQuery(`INSERT INTO "node_execution_events" ("created_at","updated_at","deleted_at","execution_project","execution_domain","execution_name","node_id","request_id","occurred_at","phase") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`)
	nodeExecEventRepo := NewNodeExecutionEventRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := nodeExecEventRepo.Create(context.Background(), models.NodeExecutionEvent{
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
	})
	assert.NoError(t, err)
	assert.True(t, nodeExecutionEventQuery.Triggered)
}
