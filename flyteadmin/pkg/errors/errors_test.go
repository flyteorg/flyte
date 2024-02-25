package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestGrpcStatusError(t *testing.T) {

	msg := "some error"
	curPhase := "some phase"
	statusErr := NewAlreadyInTerminalStateError(context.Background(), msg, curPhase)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	assert.Equal(t, "some error", s.Message())

	details, ok := s.Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestNewIncompatibleClusterError(t *testing.T) {
	errorMsg := "foo"
	cluster := "C1"
	statusErr := NewIncompatibleClusterError(context.Background(), errorMsg, cluster)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	assert.Equal(t, errorMsg, s.Message())

	details, ok := s.Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_IncompatibleCluster)
	assert.True(t, ok)
}

func TestNewTaskExistsDifferentStructureError(t *testing.T) {
	t := &admin.TaskCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "testProj",
			Domain:       "domain",
			Name:         "name",
			Version:      "ver",
		},
	}
	statusErr := NewTaskExistsDifferentStructureError(context.Background(), t)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "task with different structure already exists", s.Message())

	details, ok := s.Details()[0].(*admin.CreateTaskFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.CreateTaskFailureReason_ExistsDifferentStructure)
	assert.True(t, ok)
}

func TestNewTaskExistsIdenticalStructureError(t *testing.T) {
	t := &admin.TaskCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "testProj",
			Domain:       "domain",
			Name:         "name",
			Version:      "ver",
		},
	}
	statusErr := NewTaskExistsIdenticalStructureError(context.Background(), t)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, s.Code())
	assert.Equal(t, "task with identical structure already exists", s.Message())

	details, ok := s.Details()[0].(*admin.CreateTaskFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.CreateTaskFailureReason_ExistsIdenticalStructure)
	assert.True(t, ok)
}

func TestNewWorkflowExistsDifferentStructureError(t *testing.T) {
	wf := &admin.WorkflowCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "testProj",
			Domain:       "domain",
			Name:         "name",
			Version:      "ver",
		},
	}
	statusErr := NewWorkflowExistsDifferentStructureError(context.Background(), wf)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "workflow with different structure already exists", s.Message())

	details, ok := s.Details()[0].(*admin.CreateWorkflowFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.CreateWorkflowFailureReason_ExistsDifferentStructure)
	assert.True(t, ok)
}

func TestNewWorkflowExistsIdenticalStructureError(t *testing.T) {
	wf := &admin.WorkflowCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "testProj",
			Domain:       "domain",
			Name:         "name",
			Version:      "ver",
		},
	}
	statusErr := NewWorkflowExistsIdenticalStructureError(context.Background(), wf)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, s.Code())
	assert.Equal(t, "workflow with identical structure already exists", s.Message())

	details, ok := s.Details()[0].(*admin.CreateWorkflowFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.CreateWorkflowFailureReason_ExistsIdenticalStructure)
	assert.True(t, ok)
}

func TestIsDoesNotExistError(t *testing.T) {
	assert.True(t, IsDoesNotExistError(NewFlyteAdminError(codes.NotFound, "foo")))
}

func TestIsNotDoesNotExistError(t *testing.T) {
	assert.False(t, IsDoesNotExistError(NewFlyteAdminError(codes.Canceled, "foo")))
}

func TestIsNotDoesNotExistErrorBecauseOfNoneAdminError(t *testing.T) {
	assert.False(t, IsDoesNotExistError(errors.New("foo")))
}
