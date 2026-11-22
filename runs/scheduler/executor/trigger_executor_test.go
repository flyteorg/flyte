package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/service"
)

type recordingRun struct {
	service.Run
	req      *workflow.CreateRunRequest
	identity *common.EnrichedIdentity
	err      error
}

func (r *recordingRun) CreateRun(_ context.Context, req *workflow.CreateRunRequest, identity *common.EnrichedIdentity) (*workflow.CreateRunResponse, error) {
	r.req, r.identity = req, identity
	return nil, r.err
}

func TestExecuteCreatesScheduledRun(t *testing.T) {
	runs := &recordingRun{}
	exec := NewTriggerExecutor(runs, 10, 2)
	trigger := &models.Trigger{Project: "p", Domain: "d", Name: "nightly", TaskName: "job"}
	at := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)

	require.NoError(t, exec.Execute(t.Context(), trigger, at))
	req := runs.req
	require.Nil(t, runs.identity)
	require.Equal(t, workflow.RunSource_RUN_SOURCE_SCHEDULE_TRIGGER, req.GetSource())
	require.Equal(t, at, req.GetRunStartTime().AsTime())
	require.Equal(t, "nightly", req.GetTriggerName().GetName())
	require.Equal(t, "job", req.GetTriggerName().GetTaskName())
	require.NotEmpty(t, req.GetRunId().GetName())

	runs.err = errors.New("run creation failed")
	require.ErrorIs(t, exec.Execute(t.Context(), trigger, at), runs.err)
	require.True(t, proto.Equal(req, runs.req))
}
