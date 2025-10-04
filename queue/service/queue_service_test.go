package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// MockRepository is a mock implementation of repository.Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockRepository) AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason string) error {
	args := m.Called(ctx, runID, reason)
	return args.Error(0)
}

func (m *MockRepository) AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason string) error {
	args := m.Called(ctx, actionID, reason)
	return args.Error(0)
}

func (m *MockRepository) GetQueuedActions(ctx context.Context, limit int) (interface{}, error) {
	args := m.Called(ctx, limit)
	return args.Get(0), args.Error(1)
}

func (m *MockRepository) MarkAsProcessing(ctx context.Context, actionID *common.ActionIdentifier) error {
	args := m.Called(ctx, actionID)
	return args.Error(0)
}

func (m *MockRepository) MarkAsCompleted(ctx context.Context, actionID *common.ActionIdentifier) error {
	args := m.Called(ctx, actionID)
	return args.Error(0)
}

func (m *MockRepository) MarkAsFailed(ctx context.Context, actionID *common.ActionIdentifier, errorMsg string) error {
	args := m.Called(ctx, actionID, errorMsg)
	return args.Error(0)
}

func TestEnqueueAction(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewQueueService(mockRepo)

	req := &workflow.EnqueueActionRequest{
		ActionId: &common.ActionIdentifier{
			RunId: &common.RunIdentifier{
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
					Template: &task.TaskTemplate{
						Target: &task.TaskTemplate_Container{
							Container: &task.ContainerTask{
								Image: "alpine:latest",
								Args:  []string{"echo", "hello"},
							},
						},
					},
				},
			},
		},
	}

	mockRepo.On("EnqueueAction", mock.Anything, req).Return(nil)

	connectReq := connect.NewRequest(req)
	resp, err := svc.EnqueueAction(context.Background(), connectReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAbortQueuedRun(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewQueueService(mockRepo)

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

	mockRepo.On("AbortQueuedRun", mock.Anything, runID, reason).Return(nil)

	connectReq := connect.NewRequest(req)
	resp, err := svc.AbortQueuedRun(context.Background(), connectReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestAbortQueuedAction(t *testing.T) {
	mockRepo := new(MockRepository)
	svc := NewQueueService(mockRepo)

	actionID := &common.ActionIdentifier{
		RunId: &common.RunIdentifier{
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

	mockRepo.On("AbortQueuedAction", mock.Anything, actionID, reason).Return(nil)

	connectReq := connect.NewRequest(req)
	resp, err := svc.AbortQueuedAction(context.Background(), connectReq)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}
