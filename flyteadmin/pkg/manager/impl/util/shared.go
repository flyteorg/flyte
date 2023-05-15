// Shared method implementations.
package util

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

func GetExecutionName(request admin.ExecutionCreateRequest) string {
	if request.Name != "" {
		return request.Name
	}
	return common.GetExecutionName(time.Now().UnixNano())
}

func GetTask(ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (
	*admin.Task, error) {
	taskModel, err := GetTaskModel(ctx, repo, &identifier)
	if err != nil {
		return nil, err
	}
	task, err := transformers.FromTaskModel(*taskModel)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform task model for identifier [%+v] with err: %v", identifier, err)
		return nil, err
	}
	return &task, nil
}

func GetWorkflowModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (models.Workflow, error) {
	workflowModel, err := (repo).WorkflowRepo().Get(ctx, repoInterfaces.Identifier{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
		Version: identifier.Version,
	})
	if err != nil {
		return models.Workflow{}, err
	}
	return workflowModel, nil
}

func FetchAndGetWorkflowClosure(ctx context.Context,
	store *storage.DataStore,
	remoteLocationIdentifier string) (*admin.WorkflowClosure, error) {
	closure := &admin.WorkflowClosure{}

	err := store.ReadProtobuf(ctx, storage.DataReference(remoteLocationIdentifier), closure)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"Unable to read WorkflowClosure from location %s : %v", remoteLocationIdentifier, err)
	}
	return closure, nil
}

func GetWorkflow(
	ctx context.Context,
	repo repoInterfaces.Repository,
	store *storage.DataStore,
	identifier core.Identifier) (*admin.Workflow, error) {
	workflowModel, err := GetWorkflowModel(ctx, repo, identifier)
	if err != nil {
		return nil, err
	}
	workflow, err := transformers.FromWorkflowModel(workflowModel)
	if err != nil {
		return nil, err
	}
	closure, err := FetchAndGetWorkflowClosure(ctx, store, workflowModel.RemoteClosureIdentifier)
	if err != nil {
		return nil, err
	}
	closure.CreatedAt = workflow.Closure.CreatedAt
	workflow.Closure = closure
	return &workflow, nil
}

func GetLaunchPlanModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (models.LaunchPlan, error) {
	launchPlanModel, err := (repo).LaunchPlanRepo().Get(ctx, repoInterfaces.Identifier{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
		Version: identifier.Version,
	})
	if err != nil {
		return models.LaunchPlan{}, err
	}
	return launchPlanModel, nil
}

func GetLaunchPlan(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (*admin.LaunchPlan, error) {
	launchPlanModel, err := GetLaunchPlanModel(ctx, repo, identifier)
	if err != nil {
		return nil, err
	}
	return transformers.FromLaunchPlanModel(launchPlanModel)
}

func GetNamedEntityModel(
	ctx context.Context, repo repoInterfaces.Repository, resourceType core.ResourceType, identifier admin.NamedEntityIdentifier) (models.NamedEntity, error) {
	metadataModel, err := (repo).NamedEntityRepo().Get(ctx, repoInterfaces.GetNamedEntityInput{
		ResourceType: resourceType,
		Project:      identifier.Project,
		Domain:       identifier.Domain,
		Name:         identifier.Name,
	})
	if err != nil {
		return models.NamedEntity{}, err
	}
	return metadataModel, nil
}

func GetNamedEntity(
	ctx context.Context, repo repoInterfaces.Repository, resourceType core.ResourceType, identifier admin.NamedEntityIdentifier) (*admin.NamedEntity, error) {
	metadataModel, err := GetNamedEntityModel(ctx, repo, resourceType, identifier)
	if err != nil {
		return nil, err
	}
	metadata := transformers.FromNamedEntityModel(metadataModel)
	return &metadata, nil
}

func GetDescriptionEntityModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (models.DescriptionEntity, error) {
	descriptionEntityModel, err := (repo).DescriptionEntityRepo().Get(ctx, repoInterfaces.GetDescriptionEntityInput{
		ResourceType: identifier.ResourceType,
		Project:      identifier.Project,
		Domain:       identifier.Domain,
		Name:         identifier.Name,
		Version:      identifier.Version,
	})
	if err != nil {
		return models.DescriptionEntity{}, err
	}
	return descriptionEntityModel, nil
}

func GetDescriptionEntity(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.Identifier) (*admin.DescriptionEntity, error) {
	descriptionEntityModel, err := GetDescriptionEntityModel(ctx, repo, identifier)
	if err != nil {
		logger.Errorf(ctx, "Failed to get description entity [%+v]: %v", identifier, err)
		return nil, err
	}
	descriptionEntity, err := transformers.FromDescriptionEntityModel(descriptionEntityModel)
	if err != nil {
		logger.Errorf(ctx, "Failed to unmarshal description entity [%+v]: %v", descriptionEntityModel, err)
		return nil, err
	}
	return descriptionEntity, nil
}

// Returns the set of filters necessary to query launch plan models to find the active version of a launch plan
func GetActiveLaunchPlanVersionFilters(project, domain, name string) ([]common.InlineFilter, error) {
	projectFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Project, project)
	if err != nil {
		return nil, err
	}
	domainFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Domain, domain)
	if err != nil {
		return nil, err
	}
	nameFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Name, name)
	if err != nil {
		return nil, err
	}
	activeFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.State, int32(admin.LaunchPlanState_ACTIVE))
	if err != nil {
		return nil, err
	}
	return []common.InlineFilter{projectFilter, domainFilter, nameFilter, activeFilter}, nil
}

// Returns the set of filters necessary to query launch plan models to find the active version of a launch plan
func ListActiveLaunchPlanVersionsFilters(project, domain string) ([]common.InlineFilter, error) {
	projectFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Project, project)
	if err != nil {
		return nil, err
	}
	domainFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.Domain, domain)
	if err != nil {
		return nil, err
	}
	activeFilter, err := common.NewSingleValueFilter(common.LaunchPlan, common.Equal, shared.State, int32(admin.LaunchPlanState_ACTIVE))
	if err != nil {
		return nil, err
	}
	return []common.InlineFilter{projectFilter, domainFilter, activeFilter}, nil
}

func GetExecutionModel(
	ctx context.Context, repo repoInterfaces.Repository, identifier core.WorkflowExecutionIdentifier) (
	*models.Execution, error) {
	executionModel, err := repo.ExecutionRepo().Get(ctx, repoInterfaces.Identifier{
		Project: identifier.Project,
		Domain:  identifier.Domain,
		Name:    identifier.Name,
	})
	if err != nil {
		return nil, err
	}
	return &executionModel, nil
}

func GetNodeExecutionModel(ctx context.Context, repo repoInterfaces.Repository, nodeExecutionIdentifier *core.NodeExecutionIdentifier) (
	*models.NodeExecution, error) {
	nodeExecutionModel, err := repo.NodeExecutionRepo().Get(ctx, repoInterfaces.NodeExecutionResource{
		NodeExecutionIdentifier: *nodeExecutionIdentifier,
	})

	if err != nil {
		return nil, err
	}
	return &nodeExecutionModel, nil
}

func GetTaskModel(ctx context.Context, repo repoInterfaces.Repository, taskIdentifier *core.Identifier) (
	*models.Task, error) {
	taskModel, err := repo.TaskRepo().Get(ctx, repoInterfaces.Identifier{
		Project: taskIdentifier.Project,
		Domain:  taskIdentifier.Domain,
		Name:    taskIdentifier.Name,
		Version: taskIdentifier.Version,
	})

	if err != nil {
		return nil, err
	}
	return &taskModel, nil
}

func GetTaskExecutionModel(
	ctx context.Context, repo repoInterfaces.Repository, taskExecutionID *core.TaskExecutionIdentifier) (*models.TaskExecution, error) {
	if err := validation.ValidateTaskExecutionIdentifier(taskExecutionID); err != nil {
		logger.Debugf(ctx, "can't get task execution with invalid identifier [%v]: %v", taskExecutionID, err)
		return nil, err
	}

	taskExecutionModel, err := repo.TaskExecutionRepo().Get(ctx, repoInterfaces.GetTaskExecutionInput{
		TaskExecutionID: *taskExecutionID,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to get task execution with id [%+v] with err %v", taskExecutionID, err)
		return nil, err
	}
	return &taskExecutionModel, nil
}

// GetMatchableResource gets matchable resource for resourceType and project - domain - workflow combination.
// Returns nil with nothing is found or return an error
func GetMatchableResource(ctx context.Context, resourceManager interfaces.ResourceInterface, resourceType admin.MatchableResource,
	project, domain, workflowName string) (*interfaces.ResourceResponse, error) {
	matchableResource, err := resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     workflowName,
		ResourceType: resourceType,
	})
	if err != nil {
		if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
			logger.Errorf(ctx, "Failed to get %v overrides in %s project %s domain %s workflow with error: %v", resourceType,
				project, domain, workflowName, err)
			return nil, err
		}
	}
	return matchableResource, nil
}

// MergeIntoExecConfig into workflowExecConfig (higher priority) from spec (lower priority) and return the
// new object with the merged changes.
// After settings project is done, can move this function back to execution manager. Currently shared with resource.
func MergeIntoExecConfig(workflowExecConfig admin.WorkflowExecutionConfig, spec shared.WorkflowExecutionConfigInterface) admin.WorkflowExecutionConfig {
	if workflowExecConfig.GetMaxParallelism() == 0 && spec.GetMaxParallelism() > 0 {
		workflowExecConfig.MaxParallelism = spec.GetMaxParallelism()
	}

	// Do a deep check on the spec in case the security context is set but to an empty object (which may be
	// the case when coming from the UI)
	if workflowExecConfig.GetSecurityContext() == nil && spec.GetSecurityContext() != nil {
		if spec.GetSecurityContext().GetRunAs() != nil &&
			(len(spec.GetSecurityContext().GetRunAs().GetK8SServiceAccount()) > 0 ||
				len(spec.GetSecurityContext().GetRunAs().GetIamRole()) > 0 ||
				len(spec.GetSecurityContext().GetRunAs().GetExecutionIdentity()) > 0) {
			workflowExecConfig.SecurityContext = spec.GetSecurityContext()
		}
	}
	// Launchplan spec has label, annotation and rawOutputDataConfig initialized with empty values.
	// Hence we do a deep check in the following conditions before assignment
	if (workflowExecConfig.GetRawOutputDataConfig() == nil ||
		len(workflowExecConfig.GetRawOutputDataConfig().GetOutputLocationPrefix()) == 0) &&
		(spec.GetRawOutputDataConfig() != nil && len(spec.GetRawOutputDataConfig().OutputLocationPrefix) > 0) {
		workflowExecConfig.RawOutputDataConfig = spec.GetRawOutputDataConfig()
	}
	if (workflowExecConfig.GetLabels() == nil || len(workflowExecConfig.GetLabels().Values) == 0) &&
		(spec.GetLabels() != nil && len(spec.GetLabels().Values) > 0) {
		workflowExecConfig.Labels = spec.GetLabels()
	}
	if (workflowExecConfig.GetAnnotations() == nil || len(workflowExecConfig.GetAnnotations().Values) == 0) &&
		(spec.GetAnnotations() != nil && len(spec.GetAnnotations().Values) > 0) {
		workflowExecConfig.Annotations = spec.GetAnnotations()
	}

	if workflowExecConfig.GetInterruptible() == nil && spec.GetInterruptible() != nil {
		workflowExecConfig.Interruptible = spec.GetInterruptible()
	}

	if !workflowExecConfig.GetOverwriteCache() && spec.GetOverwriteCache() {
		workflowExecConfig.OverwriteCache = spec.GetOverwriteCache()
	}

	if (workflowExecConfig.GetEnvs() == nil || len(workflowExecConfig.GetEnvs().Values) == 0) &&
		(spec.GetEnvs() != nil && len(spec.GetEnvs().Values) > 0) {
		workflowExecConfig.Envs = spec.GetEnvs()
	}

	return workflowExecConfig
}
