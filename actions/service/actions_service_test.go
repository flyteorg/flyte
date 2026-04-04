package service

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/v2/actions/service/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

var (
	testActionID = &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "run",
		},
		Name: "action",
	}

	testAction = &actions.Action{
		ActionId:      testActionID,
		InputUri:      "s3://bucket/input",
		RunOutputBase: "s3://bucket/output",
		Spec: &actions.Action_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{
						Type: "container",
					},
				},
			},
		},
	}
)

func TestEnqueue(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		m.EXPECT().Enqueue(mock.Anything, testAction, (*task.RunSpec)(nil)).Return(nil)

		resp, err := svc.Enqueue(context.Background(), connect.NewRequest(&actions.EnqueueRequest{
			Action: testAction,
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("client error returns internal", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		m.EXPECT().Enqueue(mock.Anything, testAction, (*task.RunSpec)(nil)).Return(errors.New("k8s error"))

		_, err := svc.Enqueue(context.Background(), connect.NewRequest(&actions.EnqueueRequest{
			Action: testAction,
		}))

		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})
}

func TestGetLatestState(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		m.EXPECT().GetState(mock.Anything, testActionID).Return(`{"status":"ok"}`, nil)

		resp, err := svc.GetLatestState(context.Background(), connect.NewRequest(&actions.GetLatestStateRequest{
			ActionId: testActionID,
			Attempt:  1,
		}))

		assert.NoError(t, err)
		assert.Equal(t, `{"status":"ok"}`, resp.Msg.State)
	})

	t.Run("client error returns not found", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		m.EXPECT().GetState(mock.Anything, testActionID).Return("", errors.New("not found"))

		_, err := svc.GetLatestState(context.Background(), connect.NewRequest(&actions.GetLatestStateRequest{
			ActionId: testActionID,
			Attempt:  1,
		}))

		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})
}

func TestUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		status := &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_SUCCEEDED}
		m.EXPECT().PutState(mock.Anything, testActionID, uint32(1), status, `{}`).Return(nil)

		resp, err := svc.Update(context.Background(), connect.NewRequest(&actions.UpdateRequest{
			ActionId: testActionID,
			Attempt:  1,
			Status:   status,
			State:    `{}`,
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("client error returns internal", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		status := &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_RUNNING}
		m.EXPECT().PutState(mock.Anything, testActionID, uint32(1), status, `{}`).Return(errors.New("write failed"))

		_, err := svc.Update(context.Background(), connect.NewRequest(&actions.UpdateRequest{
			ActionId: testActionID,
			Attempt:  1,
			Status:   status,
			State:    `{}`,
		}))

		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})
}

func TestAbort(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		reason := "user requested"
		m.EXPECT().AbortAction(mock.Anything, testActionID, &reason).Return(nil)

		resp, err := svc.Abort(context.Background(), connect.NewRequest(&actions.AbortRequest{
			ActionId: testActionID,
			Reason:   &reason,
		}))

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("client error returns internal", func(t *testing.T) {
		m := mocks.NewActionsClientInterface(t)
		svc := NewActionsService(m)

		m.EXPECT().AbortAction(mock.Anything, testActionID, (*string)(nil)).Return(errors.New("delete failed"))

		_, err := svc.Abort(context.Background(), connect.NewRequest(&actions.AbortRequest{
			ActionId: testActionID,
		}))

		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
	})
}
