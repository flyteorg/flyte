package transformers

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type CreateTaskExecutionModelInput struct {
	Request *admin.TaskExecutionEventRequest
}

func addTaskStartedState(request *admin.TaskExecutionEventRequest, taskExecutionModel *models.TaskExecution,
	closure *admin.TaskExecutionClosure) error {
	occurredAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal occurredAt with error: %v", err)
	}
	taskExecutionModel.StartedAt = &occurredAt
	closure.StartedAt = request.Event.OccurredAt
	return nil
}

func addTaskTerminalState(
	request *admin.TaskExecutionEventRequest,
	taskExecutionModel *models.TaskExecution, closure *admin.TaskExecutionClosure) error {
	if taskExecutionModel.StartedAt == nil {
		logger.Warning(context.Background(), "task execution is missing StartedAt")
	} else {
		endTime, err := ptypes.Timestamp(request.Event.OccurredAt)
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse task execution occurredAt timestamp: %v", err)
		}
		closure.StartedAt, err = ptypes.TimestampProto(*taskExecutionModel.StartedAt)
		if err != nil {
			return errors.NewFlyteAdminErrorf(
				codes.Internal, "Failed to parse task execution startedAt timestamp: %v", err)
		}
		taskExecutionModel.Duration = endTime.Sub(*taskExecutionModel.StartedAt)
		closure.Duration = ptypes.DurationProto(taskExecutionModel.Duration)
	}

	if request.Event.GetOutputUri() != "" {
		closure.OutputResult = &admin.TaskExecutionClosure_OutputUri{
			OutputUri: request.Event.GetOutputUri(),
		}
	} else if request.Event.GetError() != nil {
		closure.OutputResult = &admin.TaskExecutionClosure_Error{
			Error: request.Event.GetError(),
		}
	}
	return nil
}

func CreateTaskExecutionModel(input CreateTaskExecutionModelInput) (*models.TaskExecution, error) {
	taskExecution := &models.TaskExecution{
		TaskExecutionKey: models.TaskExecutionKey{
			TaskKey: models.TaskKey{
				Project: input.Request.Event.TaskId.Project,
				Domain:  input.Request.Event.TaskId.Domain,
				Name:    input.Request.Event.TaskId.Name,
				Version: input.Request.Event.TaskId.Version,
			},
			NodeExecutionKey: models.NodeExecutionKey{
				NodeID: input.Request.Event.ParentNodeExecutionId.NodeId,
				ExecutionKey: models.ExecutionKey{
					Project: input.Request.Event.ParentNodeExecutionId.ExecutionId.Project,
					Domain:  input.Request.Event.ParentNodeExecutionId.ExecutionId.Domain,
					Name:    input.Request.Event.ParentNodeExecutionId.ExecutionId.Name,
				},
			},
			RetryAttempt: &input.Request.Event.RetryAttempt,
		},

		Phase:        input.Request.Event.Phase.String(),
		PhaseVersion: input.Request.Event.PhaseVersion,
		InputURI:     input.Request.Event.InputUri,
	}

	closure := &admin.TaskExecutionClosure{
		Phase:      input.Request.Event.Phase,
		UpdatedAt:  input.Request.Event.OccurredAt,
		CreatedAt:  input.Request.Event.OccurredAt,
		Logs:       input.Request.Event.Logs,
		CustomInfo: input.Request.Event.CustomInfo,
	}

	eventPhase := input.Request.Event.Phase

	// Different tasks may report different phases as their first event.
	// If the first event we receive for this execution is a valid
	// non-terminal phase, mark the execution start time.
	if eventPhase == core.TaskExecution_RUNNING {
		err := addTaskStartedState(input.Request, taskExecution, closure)
		if err != nil {
			return nil, err
		}
	}

	if common.IsTaskExecutionTerminal(input.Request.Event.Phase) {
		err := addTaskTerminalState(input.Request, taskExecution, closure)
		if err != nil {
			return nil, err
		}
	}
	marshaledClosure, err := proto.Marshal(closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal task execution closure with error: %v", err)
	}

	taskExecution.Closure = marshaledClosure
	taskExecutionCreatedAt, err := ptypes.Timestamp(input.Request.Event.OccurredAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read event timestamp")
	}
	taskExecution.TaskExecutionCreatedAt = &taskExecutionCreatedAt
	taskExecution.TaskExecutionUpdatedAt = &taskExecutionCreatedAt

	return taskExecution, nil
}

func UpdateTaskExecutionModel(request *admin.TaskExecutionEventRequest, taskExecutionModel *models.TaskExecution) error {
	var taskExecutionClosure admin.TaskExecutionClosure
	err := proto.Unmarshal(taskExecutionModel.Closure, &taskExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to unmarshal task execution closure with error: %+v", err)
	}
	existingTaskPhase := taskExecutionModel.Phase
	taskExecutionModel.Phase = request.Event.Phase.String()
	taskExecutionModel.PhaseVersion = request.Event.PhaseVersion
	taskExecutionClosure.Phase = request.Event.Phase
	taskExecutionClosure.UpdatedAt = request.Event.OccurredAt
	taskExecutionClosure.Logs = request.Event.Logs
	if (existingTaskPhase == core.TaskExecution_QUEUED.String() || existingTaskPhase == core.TaskExecution_UNDEFINED.String()) && taskExecutionModel.Phase == core.TaskExecution_RUNNING.String() {
		err = addTaskStartedState(request, taskExecutionModel, &taskExecutionClosure)
		if err != nil {
			return err
		}
	}

	if common.IsTaskExecutionTerminal(request.Event.Phase) {
		err := addTaskTerminalState(request, taskExecutionModel, &taskExecutionClosure)
		if err != nil {
			return err
		}
	}
	taskExecutionClosure.CustomInfo = request.Event.CustomInfo
	marshaledClosure, err := proto.Marshal(&taskExecutionClosure)
	if err != nil {
		return errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to marshal task execution closure with error: %v", err)
	}
	taskExecutionModel.Closure = marshaledClosure
	updatedAt, err := ptypes.Timestamp(request.Event.OccurredAt)
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to parse updated at timestamp")
	}
	taskExecutionModel.TaskExecutionUpdatedAt = &updatedAt
	return nil
}

func FromTaskExecutionModel(taskExecutionModel models.TaskExecution) (*admin.TaskExecution, error) {
	var closure admin.TaskExecutionClosure
	err := proto.Unmarshal(taskExecutionModel.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal closure")
	}

	taskExecution := &admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      taskExecutionModel.TaskExecutionKey.TaskKey.Project,
				Domain:       taskExecutionModel.TaskExecutionKey.TaskKey.Domain,
				Name:         taskExecutionModel.TaskExecutionKey.TaskKey.Name,
				Version:      taskExecutionModel.TaskExecutionKey.TaskKey.Version,
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: taskExecutionModel.NodeExecutionKey.NodeID,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Project,
					Domain:  taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Domain,
					Name:    taskExecutionModel.TaskExecutionKey.NodeExecutionKey.ExecutionKey.Name,
				},
			},
			RetryAttempt: *taskExecutionModel.TaskExecutionKey.RetryAttempt,
		},
		InputUri: taskExecutionModel.InputURI,
		Closure:  &closure,
	}
	if len(taskExecutionModel.ChildNodeExecution) > 0 {
		taskExecution.IsParent = true
	}

	return taskExecution, nil
}

func FromTaskExecutionModels(taskExecutionModels []models.TaskExecution) ([]*admin.TaskExecution, error) {
	taskExecutions := make([]*admin.TaskExecution, len(taskExecutionModels))
	for idx, taskExecutionModel := range taskExecutionModels {
		taskExecution, err := FromTaskExecutionModel(taskExecutionModel)
		if err != nil {
			return nil, err
		}
		taskExecutions[idx] = taskExecution
	}
	return taskExecutions, nil
}
