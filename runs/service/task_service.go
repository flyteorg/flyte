package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	flyteTask "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
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

	// Validate request using proto validation
	if err := request.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid DeployTask request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	taskSpec := request.GetSpec()

	// TODO(nary): Add semantic validation to validate that default input types match the task interface

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

	// TODO(nary): Add triggers when the trigger service is ready
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
	request := c.Msg

	// TODO(nary): Add auth and get organization from identity context
	// org := authz.IdentityContextFromContext(ctx).Organization()
	// For now, we'll skip org-level filtering

	// Convert request to ListResourceInput
	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TaskColumns)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Apply scope filters based on request
	switch scope := request.GetScopeBy().(type) {
	case *task.ListTasksRequest_Org:
		// Filter by organization
		orgFilter := impl.NewOrgFilter(scope.Org)
		listInput.ScopeByFilter = orgFilter
	case *task.ListTasksRequest_ProjectId:
		// Filter by project (org + project + domain)
		projectFilter := impl.NewProjectIdFilter(scope.ProjectId)
		listInput.ScopeByFilter = projectFilter
	}

	// Handle known filters
	if request.GetKnownFilters() != nil {
		for _, knownFilter := range request.GetKnownFilters() {
			switch filterType := knownFilter.GetFilterBy().(type) {
			case *task.ListTasksRequest_KnownFilter_DeployedBy:
				deployedByFilter := impl.NewDeployedByFilter(filterType.DeployedBy)
				if listInput.Filter == nil {
					listInput.Filter = deployedByFilter
				} else {
					listInput.Filter = listInput.Filter.And(deployedByFilter)
				}
			}
		}
	}

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
	request := c.Msg

	// Convert request to ListResourceInput
	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TaskVersionColumns)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Add task name filter
	taskNameFilter := impl.NewTaskNameFilter(request.GetTaskName())
	listInput = listInput.WithFilter(taskNameFilter)

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
