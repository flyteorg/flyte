package gormimpl

import (
	"context"
	"testing"
	"time"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestCreateExecutionEvent(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	executionEventQuery := GlobalMock.NewMock()
	executionEventQuery.WithQuery(`INSERT INTO "execution_events" ("created_at","updated_at","deleted_at",` +
		`"execution_project","execution_domain","execution_name","request_id","occurred_at","phase") VALUES ` +
		`(?,?,?,?,?,?,?,?,?)`)
	execEventRepo := NewExecutionEventRepo(GetDbForTest(t), errors.NewTestErrorTransformer(), mockScope.NewTestScope())
	err := execEventRepo.Create(context.Background(), models.ExecutionEvent{
		RequestID: "request id 1",
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "1",
		},
		OccurredAt: time.Now(),
		Phase:      core.WorkflowExecution_SUCCEEDED.String(),
	})
	assert.NoError(t, err)
	assert.True(t, executionEventQuery.Triggered)
}
