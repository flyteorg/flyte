package executor

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	workflowMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestExecuteCallsInjectedRunService(t *testing.T) {
	ctx := t.Context()
	runs := workflowMocks.NewRunServiceHandler(t)
	trigger := &models.Trigger{Project: "p", Domain: "d", TaskName: "job", Name: "nightly"}
	fireTime := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	runs.EXPECT().CreateRun(ctx, mock.Anything).
		Run(func(_ context.Context, req *connect.Request[workflow.CreateRunRequest]) {
			require.Equal(t, runName(trigger, fireTime), req.Msg.GetRunId().GetName())
			require.Equal(t, trigger.Name, req.Msg.GetTriggerName().GetName())
			require.Equal(t, workflow.RunSource_RUN_SOURCE_SCHEDULE_TRIGGER, req.Msg.GetSource())
			require.Equal(t, fireTime, req.Msg.GetRunStartTime().AsTime())
		}).Return(connect.NewResponse(&workflow.CreateRunResponse{}), nil).Once()

	require.NoError(t, NewTriggerExecutor(runs, 1, 1).Execute(ctx, trigger, fireTime))
}
