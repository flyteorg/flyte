package api

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/code"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

func TestRecordActionReturnsRepositoryFailureInBody(t *testing.T) {
	t.Cleanup(func() { cleanupTestDB(t) })

	ctx := context.Background()
	require.NoError(t, testDB.PingContext(ctx))
	require.NoError(t, renameActionsTable(ctx, "actions", "actions_unavailable"))
	t.Cleanup(func() {
		require.NoError(t, renameActionsTable(context.Background(), "actions_unavailable", "actions"))
	})

	client := workflowconnect.NewInternalRunServiceClient(newClient(), endpoint)
	response, err := client.RecordAction(ctx, connect.NewRequest(&workflow.RecordActionRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     testOrg,
				Project: testProject,
				Domain:  testDomain,
				Name:    "r" + uniqueString(),
			},
			Name: "record-action-db-failure",
		},
		Spec: &workflow.RecordActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{TaskTemplate: &core.TaskTemplate{Type: "python"}},
			},
		},
	}))

	require.NoError(t, err)
	require.Equal(t, int32(code.Code_INTERNAL), response.Msg.GetStatus().GetCode())
	t.Logf(
		"DB-FAILURE transportErr=%v status.code=%d status.message=%q",
		err,
		response.Msg.GetStatus().GetCode(),
		response.Msg.GetStatus().GetMessage(),
	)
}

func renameActionsTable(ctx context.Context, from string, to string) error {
	// Table identifiers cannot be bind parameters, so the names are concatenated.
	// Both callers pass string literals, so nothing here comes from user input.
	_, err := testDB.ExecContext(ctx, "ALTER TABLE "+from+" RENAME TO "+to)
	return err
}
