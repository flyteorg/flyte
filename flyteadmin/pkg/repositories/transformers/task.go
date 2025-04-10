// Handles translating gRPC request & response objects to and from repository model objects
package transformers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Transforms a TaskCreateRequest to a task model
func CreateTaskModel(
	request *admin.TaskCreateRequest,
	taskClosure *admin.TaskClosure,
	digest []byte) (models.Task, error) {
	closureBytes, err := proto.Marshal(taskClosure)
	if err != nil {
		return models.Task{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize task closure")
	}
	var taskType string
	if taskClosure.GetCompiledTask() != nil && taskClosure.GetCompiledTask().GetTemplate() != nil {
		taskType = taskClosure.GetCompiledTask().GetTemplate().GetType()
	}
	return models.Task{
		TaskKey: models.TaskKey{
			Project: request.GetId().GetProject(),
			Domain:  request.GetId().GetDomain(),
			Name:    request.GetId().GetName(),
			Version: request.GetId().GetVersion(),
		},
		Closure: closureBytes,
		Digest:  digest,
		Type:    taskType,
	}, nil
}

func FromTaskModel(taskModel models.Task) (admin.Task, error) {
	taskClosure := &admin.TaskClosure{}
	err := proto.Unmarshal(taskModel.Closure, taskClosure)
	if err != nil {
		return admin.Task{}, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal closure")
	}

	createdAt := timestamppb.New(taskModel.CreatedAt)
	taskClosure.CreatedAt = createdAt
	id := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      taskModel.Project,
		Domain:       taskModel.Domain,
		Name:         taskModel.Name,
		Version:      taskModel.Version,
	}
	return admin.Task{
		Id:               &id,
		Closure:          taskClosure,
		ShortDescription: taskModel.ShortDescription,
	}, nil
}

func FromTaskModels(taskModels []models.Task) ([]*admin.Task, error) {
	tasks := make([]*admin.Task, len(taskModels))
	for idx, taskModel := range taskModels {
		task, err := FromTaskModel(taskModel)
		if err != nil {
			return nil, err
		}
		tasks[idx] = &task
	}
	return tasks, nil
}

func FromTaskModelsToIdentifiers(taskModels []models.Task) []*admin.NamedEntityIdentifier {
	ids := make([]*admin.NamedEntityIdentifier, len(taskModels))
	for i, taskModel := range taskModels {
		ids[i] = &admin.NamedEntityIdentifier{
			Project: taskModel.Project,
			Domain:  taskModel.Domain,
			Name:    taskModel.Name,
		}
	}
	return ids
}
