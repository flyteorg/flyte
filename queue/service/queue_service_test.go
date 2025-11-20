package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// MockQueueClient is a mock implementation of k8s.QueueClient
type MockQueueClient struct {
	mock.Mock
}

func (m *MockQueueClient) EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockQueueClient) AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason *string) error {
	args := m.Called(ctx, runID, reason)
	return args.Error(0)
}

func (m *MockQueueClient) AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason *string) error {
	args := m.Called(ctx, actionID, reason)
	return args.Error(0)
}

func TestEnqueueAction(t *testing.T) {
	mockClient := new(MockQueueClient)
	svc := NewQueueServiceWithClient(mockClient)

	req := &workflow.EnqueueActionRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     "test-org",
				Project: "test-project",
				Domain:  "test-domain",
				Name:    "test-run",
			},
			Name: "test-action",
		},
		InputUri:      "s3://bucket/input",
		RunOutputBase: "s3://bucket/output",
		Spec: &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{
						Type: "container",
						Target: &core.TaskTemplate_Container{
							Container: &core.Container{
								Image: "alpine:latest",
								Args:  []string{"echo", "hello"},
							},
						},
					},
				},
			},
		},
	}

	mockClient.On("EnqueueAction", mock.Anything, req).Return(nil)

	connectReq := connect.NewRequest(req)
	resp, err := svc.EnqueueAction(context.Background(), connectReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestAbortQueuedRun(t *testing.T) {
	mockClient := new(MockQueueClient)
	svc := NewQueueServiceWithClient(mockClient)

	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "test-run",
	}

	reason := "test abort"
	req := &workflow.AbortQueuedRunRequest{
		RunId:  runID,
		Reason: &reason,
	}

	mockClient.On("AbortQueuedRun", mock.Anything, runID, &reason).Return(nil)

	connectReq := connect.NewRequest(req)
	resp, err := svc.AbortQueuedRun(context.Background(), connectReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockClient.AssertExpectations(t)
}

func TestAbortQueuedAction(t *testing.T) {
	mockClient := new(MockQueueClient)
	svc := NewQueueServiceWithClient(mockClient)

	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "test-org",
			Project: "test-project",
			Domain:  "test-domain",
			Name:    "test-run",
		},
		Name: "test-action",
	}

	reason := "test abort"
	req := &workflow.AbortQueuedActionRequest{
		ActionId: actionID,
		Reason:   &reason,
	}

	mockClient.On("AbortQueuedAction", mock.Anything, actionID, &reason).Return(nil)

	connectReq := connect.NewRequest(req)
	resp, err := svc.AbortQueuedAction(context.Background(), connectReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockClient.AssertExpectations(t)
}
