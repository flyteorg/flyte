package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	flyteTask "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
)

type taskService struct {
	taskconnect.UnimplementedTaskServiceHandler
	db interfaces.Repository
}

func NewTaskService(repo interfaces.Repository) taskconnect.TaskServiceHandler {
	return &taskService{
		db: repo,
	}
}

func (s *taskService) DeployTask(ctx context.Context, c *connect.Request[task.DeployTaskRequest]) (*connect.Response[task.DeployTaskResponse], error) {
	request := c.Msg
	taskSpec := request.GetSpec()

	if err := validateDefaultInputsAgainstSpecInterface(taskSpec); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Truncate fields that exceed maximum length
	if env := taskSpec.GetEnvironment(); env != nil {
		env.Description = truncateShortDescription(env.GetDescription())
	}

	if doc := taskSpec.GetDocumentation(); doc != nil {
		doc.ShortDescription = truncateShortDescription(doc.GetShortDescription())
		doc.LongDescription = truncateLongDescription(doc.GetLongDescription())
	}

	taskId := request.GetTaskId()
	taskModel, err := transformers.NewTaskModel(ctx, taskId, taskSpec)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// TODO(nary): Add triggers back when the trigger service is ready
	if len(request.GetTriggers()) > 0 {
		logger.Infof(ctx, "Triggers currently not supported")
	}

	taskModel.TotalTriggers = uint32(len(request.GetTriggers()))

	err = s.db.TaskRepo().CreateTask(ctx, taskModel)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&task.DeployTaskResponse{}), nil
}

// TODO(nary): Add back type validation after adding validators and literals packages
func validateDefaultInputsAgainstSpecInterface(spec *task.TaskSpec) error {
	// Validation temporarily disabled - need validators.AreTypesCastable and literals.CoreToFlyteLiteralType
	// variableMap := spec.GetTaskTemplate().GetInterface().GetInputs().GetVariables()
	// for _, namedParam := range spec.GetDefaultInputs() {
	// 	variable, ok := variableMap[namedParam.GetName()]
	// 	if !ok {
	// 		return fmt.Errorf("invalid default input: %s", namedParam.GetName())
	// 	}
	// 	param := namedParam.GetParameter()
	// 	if param.GetDefault() != nil {
	// 		defVarType := param.GetVar().GetType()
	// 		if !AreTypesCastableV2(defVarType, variable.GetType()) {
	// 			return fmt.Errorf("invalid default_input wrong type %s, expected %s, got %s instead",
	// 				namedParam.GetName(), variable.GetType(), defVarType)
	// 		}
	// 	}
	// }
	return nil
}

func (s *taskService) GetTaskDetails(ctx context.Context, c *connect.Request[task.GetTaskDetailsRequest]) (*connect.Response[task.GetTaskDetailsResponse], error) {
	request := c.Msg
	model, err := s.db.TaskRepo().GetTask(ctx, transformers.ToTaskKey(request.GetTaskId()))
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	// TODO(nary): Add identity enrichment after adding auth
	tasks, err := transformers.TaskModelsToTaskDetailsWithoutIdentity(ctx, []*models.Task{model})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&flyteTask.GetTaskDetailsResponse{
		Details: tasks[0],
	}), nil
}

// ListTasks lists tasks based on the provided request. This implicitly returns a list that is sorted by project, domain, name in descending order.
// Thus there is no need to add a sort by name clause to this request.
func (s *taskService) ListTasks(ctx context.Context, c *connect.Request[task.ListTasksRequest]) (*connect.Response[task.ListTasksResponse], error) {
	_ = c.Msg // TODO(nary): Use request after adding filter support

	// TODO(nary): Add auth and get organization from identity context
	// org := authz.IdentityContextFromContext(ctx).Organization()

	// TODO(nary): Add filter and scope handling after adding requests package
	listInput := interfaces.ListResourceInput{
		Limit:  50, // Default limit
		Offset: 0,
	}

	// Apply scope filters based on request
	// TODO(nary): Implement proper filter conversion from request

	taskResWithCounts, err := s.db.TaskRepo().ListTasks(ctx, listInput)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// TODO(nary): Extract task names and get latest runs after adding ActionRepo.GetLatestRunByTasks
	// For now, return tasks without latest run information
	var latestRuns map[models.TaskName]*models.Action

	// TODO(nary): Add identity enrichment after adding auth
	tasks, taskMetadata, err := transformers.TaskListResultToTasksAndMetadata(
		ctx, taskResWithCounts, latestRuns, nil, false, false)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	var token string
	if len(tasks) > 0 && len(tasks) == int(listInput.Limit) {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}

	return connect.NewResponse(&task.ListTasksResponse{
		Tasks:    tasks,
		Token:    token,
		Metadata: taskMetadata,
	}), nil
}

// ListVersions lists versions based on the provided request. This implicitly returns a list that is sorted by deployedAt in descending order.
func (s *taskService) ListVersions(ctx context.Context, c *connect.Request[task.ListVersionsRequest]) (*connect.Response[task.ListVersionsResponse], error) {
	_ = c.Msg // TODO(nary): Use request after adding filter support

	// TODO(nary): Add proper filter conversion from request after adding requests package
	listInput := interfaces.ListResourceInput{
		Limit:  50, // Default limit
		Offset: 0,
	}

	// TODO(nary): Add task name filter after adding filter packages
	// taskNameFilter := v2.NewTaskNameFilter(request.GetTaskName())

	versionModels, err := s.db.TaskRepo().ListVersions(ctx, listInput)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	versions := transformers.VersionModelsToVersionResponses(versionModels)

	// Generate pagination token
	var token string
	if len(versions) > 0 && len(versions) == int(listInput.Limit) {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}

	return connect.NewResponse(&task.ListVersionsResponse{
		Versions: versions,
		Token:    token,
	}), nil
}
