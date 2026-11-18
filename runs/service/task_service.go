package service

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	commonpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/runs/repository/impl"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/repository/transformers"
	"github.com/flyteorg/flyte/v2/runs/schedule"
)

// Task provides transport-independent task operations.
type Task interface {
	DeployTask(context.Context, *task.DeployTaskRequest) (*task.DeployTaskResponse, error)
	GetTaskDetails(context.Context, *task.GetTaskDetailsRequest) (*task.GetTaskDetailsResponse, error)
	ListTasks(context.Context, *task.ListTasksRequest) (*task.ListTasksResponse, error)
	ListVersions(context.Context, *task.ListVersionsRequest) (*task.ListVersionsResponse, error)
}

// TaskService manages task deployments and version lookups.
type TaskService struct {
	tasks    interfaces.TaskRepo
	projects *ProjectService
}

// NewTaskService constructs a task service backed by the given repository.
func NewTaskService(tasks interfaces.TaskRepo, projects *ProjectService) *TaskService {
	return &TaskService{
		tasks:    tasks,
		projects: projects,
	}
}

func (s *TaskService) DeployTask(ctx context.Context, request *task.DeployTaskRequest) (*task.DeployTaskResponse, error) {
	if err := request.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid DeployTask request: %v", err)
		return nil, ErrInvalidArgument.wrap(err)
	}

	if err := s.projects.ValidateProjectExists(ctx, request.GetTaskId().GetProject()); err != nil {
		return nil, err
	}

	taskSpec := request.GetSpec()

	// Truncate fields that exceed maximum length.
	if env := taskSpec.GetEnvironment(); env != nil {
		env.Description = truncateShortDescription(env.GetDescription())
	}
	if doc := taskSpec.GetDocumentation(); doc != nil {
		doc.ShortDescription = truncateShortDescription(doc.GetShortDescription())
		doc.LongDescription = truncateLongDescription(doc.GetLongDescription())
	}

	taskId := request.GetTaskId()

	// Build trigger models before writing anything, so validation failures are
	// caught before we touch the DB (mirrors close-source TaskService.DeployTask).
	triggerModels, err := buildTriggerModels(ctx, taskId, request.GetTriggers())
	if err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}

	taskModel, err := transformers.NewTaskModel(ctx, taskId, taskSpec)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}
	taskModel.TotalTriggers = uint32(len(request.GetTriggers()))

	// Single transaction: upsert task + insert trigger revisions + refresh task meta.
	if err := s.tasks.CreateTask(ctx, taskModel, triggerModels); err != nil {
		return nil, ErrInternal.wrap(err)
	}

	return &task.DeployTaskResponse{}, nil
}

// buildTriggerModels converts the embedded TaskTrigger list into Trigger revision models
// ready for persistence. Validation (e.g. cron expression) is performed here so errors
// surface before any DB write occurs.
func buildTriggerModels(
	ctx context.Context,
	taskId *task.TaskIdentifier,
	taskTriggers []*task.TaskTrigger,
) ([]*models.Trigger, error) {
	triggerModels := make([]*models.Trigger, 0, len(taskTriggers))
	for _, tt := range taskTriggers {
		// Validate cron expression up-front.
		if err := schedule.ValidateCronExpression(tt.GetName(), tt.GetAutomationSpec()); err != nil {
			return nil, err
		}

		spec := &triggerpb.TriggerSpec{
			Active:      tt.GetSpec().GetActive(),
			TaskVersion: taskId.GetVersion(),
			Description: truncateShortDescription(tt.GetSpec().GetDescription()),
			RunSpec:     tt.GetSpec().GetRunSpec(),
		}
		// Carry over whichever input form the request used — inline literals or an offloaded URI.
		if offloaded := tt.GetSpec().GetOffloadedInputData(); offloaded != nil {
			spec.InputWrapper = &triggerpb.TriggerSpec_OffloadedInputData{OffloadedInputData: offloaded}
		} else if inputs := tt.GetSpec().GetInputs(); inputs != nil {
			spec.InputWrapper = &triggerpb.TriggerSpec_Inputs{Inputs: inputs}
		}
		id := &commonpb.TriggerIdentifier{
			Name: &commonpb.TriggerName{
				Project:  taskId.GetProject(),
				Domain:   taskId.GetDomain(),
				TaskName: taskId.GetName(),
				Name:     tt.GetName(),
			},
			// Revision 0: repo will compute the next revision inside the transaction.
			Revision: 0,
		}
		m, err := transformers.NewTriggerModel(ctx, id, spec, tt.GetAutomationSpec())
		if err != nil {
			logger.Errorf(ctx, "failed to build trigger model for %q: %v", tt.GetName(), err)
			return nil, fmt.Errorf("failed to build trigger model for %q: %w", tt.GetName(), err)
		}
		triggerModels = append(triggerModels, m)
	}
	return triggerModels, nil
}

func (s *TaskService) GetTaskDetails(ctx context.Context, request *task.GetTaskDetailsRequest) (*task.GetTaskDetailsResponse, error) {
	model, err := s.tasks.GetTask(ctx, transformers.ToTaskKey(request.GetTaskId()))
	if err != nil {
		return nil, ErrNotFound.wrap(err)
	}

	// TODO(nary): Add identity enrichment after adding auth
	tasks, err := transformers.TaskModelsToTaskDetailsWithoutIdentity(ctx, []*models.Task{model})
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	return &task.GetTaskDetailsResponse{
		Details: tasks[0],
	}, nil
}

// ListTasks lists tasks based on the provided request. This implicitly returns a list that is sorted by project, domain, name in descending order.
// Thus there is no need to add a sort by name clause to this request.
func (s *TaskService) ListTasks(ctx context.Context, request *task.ListTasksRequest) (*task.ListTasksResponse, error) {
	// TODO(nary): Add auth and get organization from identity context
	// org := authz.IdentityContextFromContext(ctx).Organization()
	// For now, we'll skip org-level filtering

	// Convert request to ListResourceInput
	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TaskColumns)
	if err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}

	// Apply scope filters based on request
	switch scope := request.GetScopeBy().(type) {
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

	taskResWithCounts, err := s.tasks.ListTasks(ctx, listInput)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	// TODO(nary): Extract task names and get latest runs after adding ActionRepo.GetLatestRunByTasks
	// For now, return tasks without latest run information
	var latestRuns map[models.TaskName]*models.Action

	// TODO(nary): Add identity enrichment after adding auth
	tasks, taskMetadata, err := transformers.TaskListResultToTasksAndMetadata(
		ctx, taskResWithCounts, latestRuns, nil, false, false)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	var token string
	if len(tasks) > 0 && len(tasks) == int(listInput.Limit) {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}

	return &task.ListTasksResponse{
		Tasks:    tasks,
		Token:    token,
		Metadata: taskMetadata,
	}, nil
}

// ListVersions lists versions based on the provided request. This implicitly returns a list that is sorted by deployedAt in descending order.
func (s *TaskService) ListVersions(ctx context.Context, request *task.ListVersionsRequest) (*task.ListVersionsResponse, error) {
	// Convert request to ListResourceInput
	listInput, err := impl.NewListResourceInputFromProto(request.GetRequest(), models.TaskVersionColumns)
	if err != nil {
		return nil, ErrInvalidArgument.wrap(err)
	}

	// Add task name filter
	taskNameFilter := impl.NewTaskNameFilter(request.GetTaskName())
	listInput = listInput.WithFilter(taskNameFilter)

	versionModels, err := s.tasks.ListVersions(ctx, listInput)
	if err != nil {
		return nil, ErrInternal.wrap(err)
	}

	versions := transformers.VersionModelsToVersionResponses(versionModels)

	// Generate pagination token
	var token string
	if len(versions) > 0 && len(versions) == int(listInput.Limit) {
		token = fmt.Sprintf("%d", listInput.Offset+listInput.Limit)
	}

	return &task.ListVersionsResponse{
		Versions: versions,
		Token:    token,
	}, nil
}

// TruncateShortDescription truncates the short description to 255 characters if it exceeds the maximum length.
func truncateShortDescription(description string) string {
	if len(description) > 255 {
		return description[:255]
	}
	return description
}

// truncateLongDescription truncates the long description to 2048 characters if it exceeds the maximum length.
func truncateLongDescription(description string) string {
	if len(description) > 2048 {
		return description[:2048]
	}
	return description
}
