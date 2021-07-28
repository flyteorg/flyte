package impl

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flyteorg/flyteadmin/auth"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"

	dataInterfaces "github.com/flyteorg/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	eventWriter "github.com/flyteorg/flyteadmin/pkg/async/events/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/notifications"
	notificationInterfaces "github.com/flyteorg/flyteadmin/pkg/async/notifications/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/executions"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"

	"github.com/benbjohnson/clock"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/golang/protobuf/proto"
)

const childContainerQueueKey = "child_queue"

// Map of [project] -> map of [domain] -> stop watch
type projectDomainScopedStopWatchMap = map[string]map[string]*promutils.StopWatch

type executionSystemMetrics struct {
	Scope                    promutils.Scope
	ActiveExecutions         prometheus.Gauge
	ExecutionsCreated        prometheus.Counter
	ExecutionsTerminated     prometheus.Counter
	ExecutionEventsCreated   prometheus.Counter
	PropellerFailures        prometheus.Counter
	PublishNotificationError prometheus.Counter
	TransformerError         prometheus.Counter
	UnexpectedDataError      prometheus.Counter
	SpecSizeBytes            prometheus.Summary
	ClosureSizeBytes         prometheus.Summary
	AcceptanceDelay          prometheus.Summary
	PublishEventError        prometheus.Counter
}

type executionUserMetrics struct {
	Scope                        promutils.Scope
	ScheduledExecutionDelays     projectDomainScopedStopWatchMap
	WorkflowExecutionDurations   projectDomainScopedStopWatchMap
	WorkflowExecutionInputBytes  prometheus.Summary
	WorkflowExecutionOutputBytes prometheus.Summary
}

type ExecutionManager struct {
	db                        repositories.RepositoryInterface
	config                    runtimeInterfaces.Configuration
	storageClient             *storage.DataStore
	workflowExecutor          workflowengineInterfaces.Executor
	queueAllocator            executions.QueueAllocator
	_clock                    clock.Clock
	systemMetrics             executionSystemMetrics
	userMetrics               executionUserMetrics
	notificationClient        notificationInterfaces.Publisher
	urlData                   dataInterfaces.RemoteURLInterface
	workflowManager           interfaces.WorkflowInterface
	namedEntityManager        interfaces.NamedEntityInterface
	resourceManager           interfaces.ResourceInterface
	qualityOfServiceAllocator executions.QualityOfServiceAllocator
	eventPublisher            notificationInterfaces.Publisher
	dbEventWriter             eventWriter.WorkflowExecutionEventWriter
}

func getExecutionContext(ctx context.Context, id *core.WorkflowExecutionIdentifier) context.Context {
	ctx = contextutils.WithExecutionID(ctx, id.Name)
	return contextutils.WithProjectDomain(ctx, id.Project, id.Domain)
}

// Returns the unique string which identifies the authenticated end user (if any).
func getUser(ctx context.Context) string {
	identityContext := auth.IdentityContextFromContext(ctx)
	return identityContext.UserID()
}

func (m *ExecutionManager) populateExecutionQueue(
	ctx context.Context, identifier core.Identifier, compiledWorkflow *core.CompiledWorkflowClosure) {
	queueConfig := m.queueAllocator.GetQueue(ctx, identifier)
	for _, task := range compiledWorkflow.Tasks {
		container := task.Template.GetContainer()
		if container == nil {
			// Unrecognized target type, nothing to do
			continue
		}

		if queueConfig.DynamicQueue != "" {
			logger.Debugf(ctx, "Assigning %s as child queue for task %+v", queueConfig.DynamicQueue, task.Template.Id)
			container.Config = append(container.Config, &core.KeyValuePair{
				Key:   childContainerQueueKey,
				Value: queueConfig.DynamicQueue,
			})
		}
	}
}

func validateMapSize(maxEntries int, candidate map[string]string, candidateName string) error {
	if maxEntries == 0 {
		// Treat the max as unset
		return nil
	}
	if len(candidate) > maxEntries {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument, "%s has too many entries [%v > %v]",
			candidateName, len(candidate), maxEntries)
	}
	return nil
}

// Labels and annotations defined in the execution spec are preferred over those defined in the
// reference launch plan spec.
func (m *ExecutionManager) addLabelsAndAnnotations(requestSpec *admin.ExecutionSpec,
	partiallyPopulatedInputs *workflowengineInterfaces.ExecuteWorkflowInput) error {

	var labels map[string]string
	if requestSpec.Labels != nil && requestSpec.Labels.Values != nil {
		labels = requestSpec.Labels.Values
	} else if partiallyPopulatedInputs.Reference.Spec.Labels != nil &&
		partiallyPopulatedInputs.Reference.Spec.Labels.Values != nil {
		labels = partiallyPopulatedInputs.Reference.Spec.Labels.Values
	}

	var annotations map[string]string
	if requestSpec.Annotations != nil && requestSpec.Annotations.Values != nil {
		annotations = requestSpec.Annotations.Values
	} else if partiallyPopulatedInputs.Reference.Spec.Annotations != nil &&
		partiallyPopulatedInputs.Reference.Spec.Annotations.Values != nil {
		annotations = partiallyPopulatedInputs.Reference.Spec.Annotations.Values
	}

	err := validateMapSize(m.config.RegistrationValidationConfiguration().GetMaxLabelEntries(), labels, "Labels")
	if err != nil {
		return err
	}
	err = validateMapSize(
		m.config.RegistrationValidationConfiguration().GetMaxAnnotationEntries(), annotations, "Annotations")
	if err != nil {
		return err
	}

	partiallyPopulatedInputs.Labels = labels
	partiallyPopulatedInputs.Annotations = annotations
	return nil
}

func (m *ExecutionManager) addPluginOverrides(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	workflowName, launchPlanName string) ([]*admin.PluginOverride, error) {
	override, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      executionID.Project,
		Domain:       executionID.Domain,
		Workflow:     workflowName,
		LaunchPlan:   launchPlanName,
		ResourceType: admin.MatchableResource_PLUGIN_OVERRIDE,
	})
	if err != nil {
		ec, ok := err.(errors.FlyteAdminError)
		if !ok || ec.Code() != codes.NotFound {
			return nil, err
		}
	}
	if override != nil && override.Attributes != nil && override.Attributes.GetPluginOverrides() != nil {
		return override.Attributes.GetPluginOverrides().Overrides, nil
	}
	return nil, nil
}

func (m *ExecutionManager) offloadInputs(ctx context.Context, literalMap *core.LiteralMap, identifier *core.WorkflowExecutionIdentifier, key string) (storage.DataReference, error) {
	if literalMap == nil {
		literalMap = &core.LiteralMap{}
	}
	inputsURI, err := m.storageClient.ConstructReference(ctx, m.storageClient.GetBaseContainerFQN(ctx), shared.Metadata, identifier.Project, identifier.Domain, identifier.Name, key)
	if err != nil {
		return "", err
	}
	if err := m.storageClient.WriteProtobuf(ctx, inputsURI, storage.Options{}, literalMap); err != nil {
		return "", err
	}
	return inputsURI, nil
}

func createTaskDefaultLimits(ctx context.Context, task *core.CompiledTask,
	systemResourceLimits runtimeInterfaces.TaskResourceSet) runtimeInterfaces.TaskResourceSet {
	// The values below should never be used (deduce it from the request; request should be set by the time we get here).
	// Setting them here just in case we end up with requests not set. We are not adding to config because it would add
	// more confusion as its mostly not used.
	cpuLimit := "500m"
	memoryLimit := "500Mi"
	resourceEntries := task.Template.GetContainer().Resources.Requests
	var cpuIndex, memoryIndex = -1, -1
	for idx, entry := range resourceEntries {
		switch entry.Name {
		case core.Resources_CPU:
			cpuIndex = idx

		case core.Resources_MEMORY:
			memoryIndex = idx
		}
	}

	if cpuIndex < 0 || memoryIndex < 0 {
		logger.Errorf(ctx, "Cpu request and Memory request missing for %s", task.Template.Id)
	}

	if cpuIndex >= 0 {
		cpuLimit = resourceEntries[cpuIndex].Value
	}
	if memoryIndex >= 0 {
		memoryLimit = resourceEntries[memoryIndex].Value
	}

	taskResourceLimits := runtimeInterfaces.TaskResourceSet{CPU: cpuLimit, Memory: memoryLimit}
	// Use the limits from config
	if systemResourceLimits.CPU != "" {
		taskResourceLimits.CPU = systemResourceLimits.CPU
	}
	if systemResourceLimits.Memory != "" {
		taskResourceLimits.Memory = systemResourceLimits.Memory
	}
	if systemResourceLimits.GPU != "" {
		taskResourceLimits.GPU = systemResourceLimits.GPU
	}

	return taskResourceLimits
}

func assignResourcesIfUnset(ctx context.Context, identifier *core.Identifier,
	platformValues runtimeInterfaces.TaskResourceSet,
	resourceEntries []*core.Resources_ResourceEntry, taskResourceSpec *admin.TaskResourceSpec) []*core.Resources_ResourceEntry {
	var cpuIndex, memoryIndex = -1, -1
	for idx, entry := range resourceEntries {
		switch entry.Name {
		case core.Resources_CPU:
			cpuIndex = idx
		case core.Resources_MEMORY:
			memoryIndex = idx
		}
	}
	if cpuIndex > 0 && memoryIndex > 0 {
		// nothing to do
		return resourceEntries
	}

	if cpuIndex < 0 && platformValues.CPU != "" {
		logger.Debugf(ctx, "Setting 'cpu' for [%+v] to %s", identifier, platformValues.CPU)
		cpuValue := platformValues.CPU
		if taskResourceSpec != nil && len(taskResourceSpec.Cpu) > 0 {
			// Use the custom attributes from the database rather than the platform defaults from the application config
			cpuValue = taskResourceSpec.Cpu
		}
		cpuResource := &core.Resources_ResourceEntry{
			Name:  core.Resources_CPU,
			Value: cpuValue,
		}
		resourceEntries = append(resourceEntries, cpuResource)
	}
	if memoryIndex < 0 && platformValues.Memory != "" {
		memoryValue := platformValues.Memory
		if taskResourceSpec != nil && len(taskResourceSpec.Memory) > 0 {
			// Use the custom attributes from the database rather than the platform defaults from the application config
			memoryValue = taskResourceSpec.Memory
		}
		memoryResource := &core.Resources_ResourceEntry{
			Name:  core.Resources_MEMORY,
			Value: memoryValue,
		}
		logger.Debugf(ctx, "Setting 'memory' for [%+v] to %s", identifier, platformValues.Memory)
		resourceEntries = append(resourceEntries, memoryResource)
	}
	return resourceEntries
}

func checkTaskRequestsLessThanLimits(ctx context.Context, identifier *core.Identifier,
	taskResources *core.Resources) {
	// We choose the minimum of the platform request defaults or the limit itself for every resource request.
	// Otherwise we can find ourselves in confusing scenarios where the injected platform request defaults exceed a
	// user-specified limit
	resourceLimits := make(map[core.Resources_ResourceName]string)
	for _, resourceEntry := range taskResources.Limits {
		resourceLimits[resourceEntry.Name] = resourceEntry.Value
	}

	finalizedResourceRequests := make([]*core.Resources_ResourceEntry, 0, len(taskResources.Requests))
	for _, resourceEntry := range taskResources.Requests {
		value := resourceEntry.Value
		quantity := resource.MustParse(resourceEntry.Value)
		limitValue, ok := resourceLimits[resourceEntry.Name]
		if !ok {
			// Unexpected - at this stage both requests and limits should be populated.
			logger.Warningf(ctx, "No limit specified for [%v] resource [%s] although request was set", identifier,
				resourceEntry.Name)
			continue
		}
		if quantity.Cmp(resource.MustParse(limitValue)) == 1 {
			// The quantity is greater than the limit! Course correct below.
			logger.Infof(ctx, "Updating requested value for task [%+v] resource [%s]. Overriding to smaller limit value [%s] from original request [%s]",
				identifier, resourceEntry.Name, limitValue, value)
			value = limitValue
		}
		finalizedResourceRequests = append(finalizedResourceRequests, &core.Resources_ResourceEntry{
			Name:  resourceEntry.Name,
			Value: value,
		})
	}
	taskResources.Requests = finalizedResourceRequests
}

// Assumes input contains a compiled task with a valid container resource execConfig.
//
// Note: The system will assign a system-default value for request but for limit it will deduce it from the request
// itself => Limit := Min([Some-Multiplier X Request], System-Max). For now we are using a multiplier of 1. In
// general we recommend the users to set limits close to requests for more predictability in the system.
func (m *ExecutionManager) setCompiledTaskDefaults(ctx context.Context, task *core.CompiledTask, workflowName string) {
	if task == nil {
		logger.Warningf(ctx, "Can't set default resources for nil task.")
		return
	}
	if task.Template == nil || task.Template.GetContainer() == nil {
		// Nothing to do
		logger.Debugf(ctx, "Not setting default resources for task [%+v], no container resources found to check", task)
		return
	}

	if task.Template.GetContainer().Resources == nil {
		// In case of no resources on the container, create empty requests and limits
		// so the container will still have resources configure properly
		task.Template.GetContainer().Resources = &core.Resources{
			Requests: []*core.Resources_ResourceEntry{},
			Limits:   []*core.Resources_ResourceEntry{},
		}
	}
	resource, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      task.Template.Id.Project,
		Domain:       task.Template.Id.Domain,
		Workflow:     workflowName,
		ResourceType: admin.MatchableResource_TASK_RESOURCE,
	})

	if err != nil {
		logger.Warningf(ctx, "Failed to fetch override values when assigning task resource default values for [%+v]: %v",
			task.Template, err)
	}
	logger.Debugf(ctx, "Assigning task requested resources for [%+v]", task.Template.Id)
	var taskResourceSpec *admin.TaskResourceSpec
	if resource != nil && resource.Attributes != nil && resource.Attributes.GetTaskResourceAttributes() != nil {
		taskResourceSpec = resource.Attributes.GetTaskResourceAttributes().Defaults
	}
	task.Template.GetContainer().Resources.Requests = assignResourcesIfUnset(
		ctx, task.Template.Id, m.config.TaskResourceConfiguration().GetDefaults(), task.Template.GetContainer().Resources.Requests,
		taskResourceSpec)

	logger.Debugf(ctx, "Assigning task resource limits for [%+v]", task.Template.Id)
	if resource != nil && resource.Attributes != nil && resource.Attributes.GetTaskResourceAttributes() != nil {
		taskResourceSpec = resource.Attributes.GetTaskResourceAttributes().Limits
	}
	task.Template.GetContainer().Resources.Limits = assignResourcesIfUnset(
		ctx, task.Template.Id, createTaskDefaultLimits(ctx, task, m.config.TaskResourceConfiguration().GetLimits()), task.Template.GetContainer().Resources.Limits,
		taskResourceSpec)
	checkTaskRequestsLessThanLimits(ctx, task.Template.Id, task.Template.GetContainer().Resources)
}

// Fetches inherited execution metadata including the parent node execution db model id and the source execution model id
// as well as sets request spec metadata with the inherited principal and adjusted nesting data.
func (m *ExecutionManager) getInheritedExecMetadata(ctx context.Context, requestSpec *admin.ExecutionSpec,
	workflowExecutionID *core.WorkflowExecutionIdentifier) (parentNodeExecutionID uint, sourceExecutionID uint, err error) {
	if requestSpec.Metadata == nil || requestSpec.Metadata.ParentNodeExecution == nil {
		return parentNodeExecutionID, sourceExecutionID, nil
	}
	parentNodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, requestSpec.Metadata.ParentNodeExecution)
	if err != nil {
		logger.Errorf(ctx, "Failed to get node execution [%+v] that launched this execution [%+v] with error %v",
			requestSpec.Metadata.ParentNodeExecution, workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}

	parentNodeExecutionID = parentNodeExecutionModel.ID

	sourceExecutionModel, err := util.GetExecutionModel(ctx, m.db, *requestSpec.Metadata.ParentNodeExecution.ExecutionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get workflow execution [%+v] that launched this execution [%+v] with error %v",
			requestSpec.Metadata.ParentNodeExecution, workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}
	sourceExecutionID = sourceExecutionModel.ID
	requestSpec.Metadata.Principal = sourceExecutionModel.User
	sourceExecution, err := transformers.FromExecutionModel(*sourceExecutionModel)
	if err != nil {
		logger.Errorf(ctx, "Failed transform parent execution model for child execution [%+v] with err: %v", workflowExecutionID, err)
		return parentNodeExecutionID, sourceExecutionID, err
	}
	if sourceExecution.Spec.Metadata != nil {
		requestSpec.Metadata.Nesting = sourceExecution.Spec.Metadata.Nesting + 1
	} else {
		requestSpec.Metadata.Nesting = 1
	}
	return parentNodeExecutionID, sourceExecutionID, nil
}

// Produces execution-time attributes for workflow execution.
// Defaults to overridable execution values set in the execution create request, then looks at the launch plan values
// (if any) before defaulting to values set in the matchable resource db.
func (m *ExecutionManager) getExecutionConfig(ctx context.Context, request *admin.ExecutionCreateRequest,
	launchPlan *admin.LaunchPlan) (*admin.WorkflowExecutionConfig, error) {
	if request.Spec.MaxParallelism > 0 {
		return &admin.WorkflowExecutionConfig{
			MaxParallelism: request.Spec.MaxParallelism,
		}, nil
	}
	if launchPlan != nil && launchPlan.Spec.MaxParallelism > 0 {
		return &admin.WorkflowExecutionConfig{
			MaxParallelism: launchPlan.Spec.MaxParallelism,
		}, nil
	}

	resource, err := m.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      request.Project,
		Domain:       request.Domain,
		ResourceType: admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
	})
	if err != nil {
		if flyteAdminError, ok := err.(errors.FlyteAdminError); !ok || flyteAdminError.Code() != codes.NotFound {
			logger.Errorf(ctx, "Failed to get workflow execution config overrides with error: %v", err)
			return nil, err
		}
	}
	if resource != nil && resource.Attributes.GetWorkflowExecutionConfig() != nil {
		return resource.Attributes.GetWorkflowExecutionConfig(), nil
	}
	return nil, nil
}

func (m *ExecutionManager) launchSingleTaskExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	context.Context, *models.Execution, error) {

	taskModel, err := m.db.TaskRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: request.Spec.LaunchPlan.Project,
		Domain:  request.Spec.LaunchPlan.Domain,
		Name:    request.Spec.LaunchPlan.Name,
		Version: request.Spec.LaunchPlan.Version,
	})
	if err != nil {
		return nil, nil, err
	}
	task, err := transformers.FromTaskModel(taskModel)
	if err != nil {
		return nil, nil, err
	}

	// Prepare a skeleton workflow
	taskIdentifier := request.Spec.LaunchPlan
	workflowModel, err :=
		util.CreateOrGetWorkflowModel(ctx, request, m.db, m.workflowManager, m.namedEntityManager, taskIdentifier, &task)
	if err != nil {
		logger.Debugf(ctx, "Failed to created skeleton workflow for [%+v] with err: %v", taskIdentifier, err)
		return nil, nil, err
	}
	workflow, err := transformers.FromWorkflowModel(*workflowModel)
	if err != nil {
		return nil, nil, err
	}
	closure, err := util.FetchAndGetWorkflowClosure(ctx, m.storageClient, workflowModel.RemoteClosureIdentifier)
	if err != nil {
		return nil, nil, err
	}
	closure.CreatedAt = workflow.Closure.CreatedAt
	workflow.Closure = closure
	// Also prepare a skeleton launch plan.
	launchPlan, err := util.CreateOrGetLaunchPlan(ctx, m.db, m.config, taskIdentifier,
		workflow.Closure.CompiledWorkflow.Primary.Template.Interface, workflowModel.ID, request.Spec)
	if err != nil {
		return nil, nil, err
	}

	name := util.GetExecutionName(request)
	workflowExecutionID := core.WorkflowExecutionIdentifier{
		Project: request.Project,
		Domain:  request.Domain,
		Name:    name,
	}
	ctx = getExecutionContext(ctx, &workflowExecutionID)

	requestSpec := request.Spec
	if requestSpec.Metadata == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.Principal = getUser(ctx)

	// Get the node execution (if any) that launched this execution
	var parentNodeExecutionID uint
	var sourceExecutionID uint
	parentNodeExecutionID, sourceExecutionID, err = m.getInheritedExecMetadata(ctx, requestSpec, &workflowExecutionID)
	if err != nil {
		return nil, nil, err
	}

	// Dynamically assign task resource defaults.
	for _, task := range workflow.Closure.CompiledWorkflow.Tasks {
		m.setCompiledTaskDefaults(ctx, task, name)
	}

	// Dynamically assign execution queues.
	m.populateExecutionQueue(ctx, *workflow.Id, workflow.Closure.CompiledWorkflow)

	inputsURI, err := m.offloadInputs(ctx, request.Inputs, &workflowExecutionID, shared.Inputs)
	if err != nil {
		return nil, nil, err
	}
	userInputsURI, err := m.offloadInputs(ctx, request.Inputs, &workflowExecutionID, shared.UserInputs)
	if err != nil {
		return nil, nil, err
	}
	qualityOfService, err := m.qualityOfServiceAllocator.GetQualityOfService(ctx, executions.GetQualityOfServiceInput{
		Workflow:               &workflow,
		LaunchPlan:             launchPlan,
		ExecutionCreateRequest: &request,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to get quality of service for [%+v] with error: %v", workflowExecutionID, err)
		return nil, nil, err
	}
	executionConfig, err := m.getExecutionConfig(ctx, &request, nil)
	if err != nil {
		return nil, nil, err
	}
	executeTaskInputs := workflowengineInterfaces.ExecuteTaskInput{
		ExecutionID:     &workflowExecutionID,
		WfClosure:       *workflow.Closure.CompiledWorkflow,
		Inputs:          request.Inputs,
		ReferenceName:   taskIdentifier.Name,
		AcceptedAt:      requestedAt,
		Auth:            requestSpec.AuthRole,
		QueueingBudget:  qualityOfService.QueuingBudget,
		ExecutionConfig: executionConfig,
	}
	if requestSpec.Labels != nil {
		executeTaskInputs.Labels = requestSpec.Labels.Values
	}
	executeTaskInputs.Labels, err = m.addProjectLabels(ctx, request.Project, executeTaskInputs.Labels)
	if err != nil {
		return nil, nil, err
	}

	if requestSpec.Annotations != nil {
		executeTaskInputs.Annotations = requestSpec.Annotations.Values
	}

	overrides, err := m.addPluginOverrides(ctx, &workflowExecutionID, workflowExecutionID.Name, "")
	if err != nil {
		return nil, nil, err
	}
	if overrides != nil {
		executeTaskInputs.TaskPluginOverrides = overrides
	}

	execInfo, err := m.workflowExecutor.ExecuteTask(ctx, executeTaskInputs)
	if err != nil {
		m.systemMetrics.PropellerFailures.Inc()
		logger.Infof(ctx, "Failed to execute workflow %+v with execution id %+v and inputs %+v with err %v",
			request, workflowExecutionID, request.Inputs, err)
		return nil, nil, err
	}
	executionCreatedAt := time.Now()
	acceptanceDelay := executionCreatedAt.Sub(requestedAt)
	m.systemMetrics.AcceptanceDelay.Observe(acceptanceDelay.Seconds())

	// Request notification settings takes precedence over the launch plan settings.
	// If there is no notification in the request and DisableAll is not true, use the settings from the launch plan.
	var notificationsSettings []*admin.Notification
	if launchPlan.Spec.GetEntityMetadata() != nil {
		notificationsSettings = launchPlan.Spec.EntityMetadata.GetNotifications()
	}
	if request.Spec.GetNotifications() != nil && request.Spec.GetNotifications().Notifications != nil &&
		len(request.Spec.GetNotifications().Notifications) > 0 {
		notificationsSettings = request.Spec.GetNotifications().Notifications
	} else if request.Spec.GetDisableAll() {
		notificationsSettings = make([]*admin.Notification, 0)
	}

	executionModel, err := transformers.CreateExecutionModel(transformers.CreateExecutionModelInput{
		WorkflowExecutionID: workflowExecutionID,
		RequestSpec:         requestSpec,
		TaskID:              taskModel.ID,
		WorkflowID:          workflowModel.ID,
		// The execution is not considered running until the propeller sends a specific event saying so.
		Phase:                 core.WorkflowExecution_UNDEFINED,
		CreatedAt:             m._clock.Now(),
		Notifications:         notificationsSettings,
		WorkflowIdentifier:    workflow.Id,
		ParentNodeExecutionID: parentNodeExecutionID,
		SourceExecutionID:     sourceExecutionID,
		Cluster:               execInfo.Cluster,
		InputsURI:             inputsURI,
		UserInputsURI:         userInputsURI,
	})
	if err != nil {
		logger.Infof(ctx, "Failed to create execution model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, err
	}
	return ctx, executionModel, nil

}

func resolvePermissions(request *admin.ExecutionCreateRequest, launchPlan *admin.LaunchPlan) *admin.AuthRole {
	if request.Spec.AuthRole != nil {
		return request.Spec.AuthRole
	}

	// Set role permissions based on launch plan Auth values.
	// The branched-ness of this check is due to the presence numerous deprecated fields
	if launchPlan.Spec.GetAuthRole() != nil {
		return launchPlan.Spec.GetAuthRole()
	} else if launchPlan.GetSpec().GetAuth() != nil {
		return &admin.AuthRole{
			AssumableIamRole:         launchPlan.GetSpec().GetAuth().AssumableIamRole,
			KubernetesServiceAccount: launchPlan.GetSpec().GetAuth().KubernetesServiceAccount,
		}
	} else if len(launchPlan.GetSpec().GetRole()) > 0 {
		return &admin.AuthRole{
			AssumableIamRole: launchPlan.GetSpec().GetRole(),
		}
	}
	return &admin.AuthRole{}
}

func (m *ExecutionManager) launchExecutionAndPrepareModel(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	context.Context, *models.Execution, error) {
	err := validation.ValidateExecutionRequest(ctx, request, m.db, m.config.ApplicationConfiguration())
	if err != nil {
		logger.Debugf(ctx, "Failed to validate ExecutionCreateRequest %+v with err %v", request, err)
		return nil, nil, err
	}
	if request.Spec.LaunchPlan.ResourceType == core.ResourceType_TASK {
		logger.Debugf(ctx, "Launching single task execution with [%+v]", request.Spec.LaunchPlan)
		return m.launchSingleTaskExecution(ctx, request, requestedAt)
	}

	launchPlanModel, err := util.GetLaunchPlanModel(ctx, m.db, *request.Spec.LaunchPlan)
	if err != nil {
		logger.Debugf(ctx, "Failed to get launch plan model for ExecutionCreateRequest %+v with err %v", request, err)
		return nil, nil, err
	}
	launchPlan, err := transformers.FromLaunchPlanModel(launchPlanModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform launch plan model %+v with err %v", launchPlanModel, err)
		return nil, nil, err
	}
	executionInputs, err := validation.CheckAndFetchInputsForExecution(
		request.Inputs,
		launchPlan.Spec.FixedInputs,
		launchPlan.Closure.ExpectedInputs,
	)

	if err != nil {
		logger.Debugf(ctx, "Failed to CheckAndFetchInputsForExecution with request.Inputs: %+v"+
			"fixed inputs: %+v and expected inputs: %+v with err %v",
			request.Inputs, launchPlan.Spec.FixedInputs, launchPlan.Closure.ExpectedInputs, err)
		return nil, nil, err
	}

	workflow, err := util.GetWorkflow(ctx, m.db, m.storageClient, *launchPlan.Spec.WorkflowId)

	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow with id %+v with err %v", launchPlan.Spec.WorkflowId, err)
		return nil, nil, err
	}
	name := util.GetExecutionName(request)
	workflowExecutionID := core.WorkflowExecutionIdentifier{
		Project: request.Project,
		Domain:  request.Domain,
		Name:    name,
	}
	ctx = getExecutionContext(ctx, &workflowExecutionID)
	var requestSpec = request.Spec
	if requestSpec.Metadata == nil {
		requestSpec.Metadata = &admin.ExecutionMetadata{}
	}
	requestSpec.Metadata.Principal = getUser(ctx)

	// Get the node and parent execution (if any) that launched this execution
	var parentNodeExecutionID uint
	var sourceExecutionID uint
	parentNodeExecutionID, sourceExecutionID, err = m.getInheritedExecMetadata(ctx, requestSpec, &workflowExecutionID)
	if err != nil {
		return nil, nil, err
	}

	// Dynamically assign task resource defaults.
	for _, task := range workflow.Closure.CompiledWorkflow.Tasks {
		m.setCompiledTaskDefaults(ctx, task, name)
	}

	// Dynamically assign execution queues.
	m.populateExecutionQueue(ctx, *workflow.Id, workflow.Closure.CompiledWorkflow)

	inputsURI, err := m.offloadInputs(ctx, executionInputs, &workflowExecutionID, shared.Inputs)
	if err != nil {
		return nil, nil, err
	}
	userInputsURI, err := m.offloadInputs(ctx, request.Inputs, &workflowExecutionID, shared.UserInputs)
	if err != nil {
		return nil, nil, err
	}

	qualityOfService, err := m.qualityOfServiceAllocator.GetQualityOfService(ctx, executions.GetQualityOfServiceInput{
		Workflow:               workflow,
		LaunchPlan:             launchPlan,
		ExecutionCreateRequest: &request,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to get quality of service for [%+v] with error: %v", workflowExecutionID, err)
		return nil, nil, err
	}
	executionConfig, err := m.getExecutionConfig(ctx, &request, launchPlan)
	if err != nil {
		return nil, nil, err
	}

	// TODO: Reduce CRD size and use offloaded input URI to blob store instead.
	executeWorkflowInputs := workflowengineInterfaces.ExecuteWorkflowInput{
		ExecutionID:     &workflowExecutionID,
		WfClosure:       *workflow.Closure.CompiledWorkflow,
		Inputs:          executionInputs,
		Reference:       *launchPlan,
		AcceptedAt:      requestedAt,
		QueueingBudget:  qualityOfService.QueuingBudget,
		ExecutionConfig: executionConfig,
		Auth:            resolvePermissions(&request, launchPlan),
	}
	err = m.addLabelsAndAnnotations(request.Spec, &executeWorkflowInputs)
	if err != nil {
		return nil, nil, err
	}
	executeWorkflowInputs.Labels, err = m.addProjectLabels(ctx, request.Project, executeWorkflowInputs.Labels)
	if err != nil {
		return nil, nil, err
	}

	overrides, err := m.addPluginOverrides(ctx, &workflowExecutionID, launchPlan.GetSpec().WorkflowId.Name, launchPlan.Id.Name)
	if err != nil {
		return nil, nil, err
	}
	if overrides != nil {
		executeWorkflowInputs.TaskPluginOverrides = overrides
	}
	if request.Spec.Metadata != nil && request.Spec.Metadata.ReferenceExecution != nil &&
		request.Spec.Metadata.Mode == admin.ExecutionMetadata_RECOVERED {
		executeWorkflowInputs.RecoveryExecution = request.Spec.Metadata.ReferenceExecution
	}

	execInfo, err := m.workflowExecutor.ExecuteWorkflow(ctx, executeWorkflowInputs)
	if err != nil {
		m.systemMetrics.PropellerFailures.Inc()
		logger.Infof(ctx, "Failed to execute workflow %+v with execution id %+v and inputs %+v with err %v",
			request, workflowExecutionID, executionInputs, err)
		return nil, nil, err
	}
	executionCreatedAt := time.Now()
	acceptanceDelay := executionCreatedAt.Sub(requestedAt)
	m.systemMetrics.AcceptanceDelay.Observe(acceptanceDelay.Seconds())

	// Request notification settings takes precedence over the launch plan settings.
	// If there is no notification in the request and DisableAll is not true, use the settings from the launch plan.
	var notificationsSettings []*admin.Notification
	if launchPlan.Spec.GetEntityMetadata() != nil {
		notificationsSettings = launchPlan.Spec.EntityMetadata.GetNotifications()
	}
	if requestSpec.GetNotifications() != nil && requestSpec.GetNotifications().Notifications != nil &&
		len(requestSpec.GetNotifications().Notifications) > 0 {
		notificationsSettings = requestSpec.GetNotifications().Notifications
	} else if requestSpec.GetDisableAll() {
		notificationsSettings = make([]*admin.Notification, 0)
	}

	executionModel, err := transformers.CreateExecutionModel(transformers.CreateExecutionModelInput{
		WorkflowExecutionID: workflowExecutionID,
		RequestSpec:         requestSpec,
		LaunchPlanID:        launchPlanModel.ID,
		WorkflowID:          launchPlanModel.WorkflowID,
		// The execution is not considered running until the propeller sends a specific event saying so.
		Phase:                 core.WorkflowExecution_UNDEFINED,
		CreatedAt:             m._clock.Now(),
		Notifications:         notificationsSettings,
		WorkflowIdentifier:    workflow.Id,
		ParentNodeExecutionID: parentNodeExecutionID,
		SourceExecutionID:     sourceExecutionID,
		Cluster:               execInfo.Cluster,
		InputsURI:             inputsURI,
		UserInputsURI:         userInputsURI,
	})
	if err != nil {
		logger.Infof(ctx, "Failed to create execution model in transformer for id: [%+v] with err: %v",
			workflowExecutionID, err)
		return nil, nil, err
	}
	return ctx, executionModel, nil
}

// Inserts an execution model into the database store and emits platform metrics.
func (m *ExecutionManager) createExecutionModel(
	ctx context.Context, executionModel *models.Execution) (*core.WorkflowExecutionIdentifier, error) {
	workflowExecutionIdentifier := core.WorkflowExecutionIdentifier{
		Project: executionModel.ExecutionKey.Project,
		Domain:  executionModel.ExecutionKey.Domain,
		Name:    executionModel.ExecutionKey.Name,
	}
	err := m.db.ExecutionRepo().Create(ctx, *executionModel)
	if err != nil {
		logger.Debugf(ctx, "failed to save newly created execution [%+v] with id %+v to db with err %v",
			workflowExecutionIdentifier, workflowExecutionIdentifier, err)
		return nil, err
	}
	m.systemMetrics.ActiveExecutions.Inc()
	m.systemMetrics.ExecutionsCreated.Inc()
	m.systemMetrics.SpecSizeBytes.Observe(float64(len(executionModel.Spec)))
	m.systemMetrics.ClosureSizeBytes.Observe(float64(len(executionModel.Closure)))
	return &workflowExecutionIdentifier, nil
}

func (m *ExecutionManager) CreateExecution(
	ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	// Prior to  flyteidl v0.15.0, Inputs was held in ExecutionSpec. Ensure older clients continue to work.
	if request.Inputs == nil || len(request.Inputs.Literals) == 0 {
		request.Inputs = request.GetSpec().GetInputs()
	}
	var executionModel *models.Execution
	var err error
	ctx, executionModel, err = m.launchExecutionAndPrepareModel(ctx, request, requestedAt)
	if err != nil {
		return nil, err
	}
	workflowExecutionIdentifier, err := m.createExecutionModel(ctx, executionModel)
	if err != nil {
		return nil, err
	}
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) RelaunchExecution(
	ctx context.Context, request admin.ExecutionRelaunchRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	existingExecutionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err %v", request, err)
		return nil, err
	}
	existingExecution, err := transformers.FromExecutionModel(*existingExecutionModel)
	if err != nil {
		return nil, err
	}

	executionSpec := existingExecution.Spec
	if executionSpec.Metadata == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	var inputs *core.LiteralMap
	if len(existingExecutionModel.UserInputsURI) > 0 {
		inputs = &core.LiteralMap{}
		if err := m.storageClient.ReadProtobuf(ctx, existingExecutionModel.UserInputsURI, inputs); err != nil {
			return nil, err
		}
	} else {
		// For old data, inputs are held in the spec
		var spec admin.ExecutionSpec
		err = proto.Unmarshal(existingExecutionModel.Spec, &spec)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to unmarshal spec")
		}
		inputs = spec.Inputs
	}
	executionSpec.Metadata.Mode = admin.ExecutionMetadata_RELAUNCH
	executionSpec.Metadata.ReferenceExecution = existingExecution.Id
	var executionModel *models.Execution
	ctx, executionModel, err = m.launchExecutionAndPrepareModel(ctx, admin.ExecutionCreateRequest{
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Name,
		Spec:    executionSpec,
		Inputs:  inputs,
	}, requestedAt)
	if err != nil {
		return nil, err
	}
	executionModel.SourceExecutionID = existingExecutionModel.ID
	workflowExecutionIdentifier, err := m.createExecutionModel(ctx, executionModel)
	if err != nil {
		return nil, err
	}
	logger.Debugf(ctx, "Successfully relaunched [%+v] as [%+v]", request.Id, workflowExecutionIdentifier)
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) RecoverExecution(
	ctx context.Context, request admin.ExecutionRecoverRequest, requestedAt time.Time) (
	*admin.ExecutionCreateResponse, error) {
	existingExecutionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err %v", request, err)
		return nil, err
	}
	existingExecution, err := transformers.FromExecutionModel(*existingExecutionModel)
	if err != nil {
		return nil, err
	}

	executionSpec := existingExecution.Spec
	if executionSpec.Metadata == nil {
		executionSpec.Metadata = &admin.ExecutionMetadata{}
	}
	var inputs *core.LiteralMap
	if len(existingExecutionModel.UserInputsURI) > 0 {
		inputs = &core.LiteralMap{}
		if err := m.storageClient.ReadProtobuf(ctx, existingExecutionModel.UserInputsURI, inputs); err != nil {
			return nil, err
		}
	}
	if request.Metadata != nil {
		executionSpec.Metadata.ParentNodeExecution = request.Metadata.ParentNodeExecution
	}
	executionSpec.Metadata.Mode = admin.ExecutionMetadata_RECOVERED
	executionSpec.Metadata.ReferenceExecution = existingExecution.Id
	var executionModel *models.Execution
	ctx, executionModel, err = m.launchExecutionAndPrepareModel(ctx, admin.ExecutionCreateRequest{
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Name,
		Spec:    executionSpec,
		Inputs:  inputs,
	}, requestedAt)
	if err != nil {
		return nil, err
	}
	executionModel.SourceExecutionID = existingExecutionModel.ID
	workflowExecutionIdentifier, err := m.createExecutionModel(ctx, executionModel)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "Successfully recovered [%+v] as [%+v]", request.Id, workflowExecutionIdentifier)
	return &admin.ExecutionCreateResponse{
		Id: workflowExecutionIdentifier,
	}, nil
}

func (m *ExecutionManager) emitScheduledWorkflowMetrics(
	ctx context.Context, executionModel *models.Execution, runningEventTimeProto *timestamp.Timestamp) {
	if executionModel == nil || runningEventTimeProto == nil {
		logger.Warningf(context.Background(),
			"tried to calculate scheduled workflow execution stats with a nil execution or event time")
		return
	}
	// Find the reference launch plan to get the kickoff time argument
	execution, err := transformers.FromExecutionModel(*executionModel)
	if err != nil {
		logger.Warningf(context.Background(),
			"failed to transform execution model when emitting scheduled workflow execution stats with for "+
				"[%s/%s/%s]", executionModel.Project, executionModel.Domain, executionModel.Name)
		return
	}
	launchPlan, err := util.GetLaunchPlan(context.Background(), m.db, *execution.Spec.LaunchPlan)
	if err != nil {
		logger.Warningf(context.Background(),
			"failed to find launch plan when emitting scheduled workflow execution stats with for "+
				"execution: [%+v] and launch plan [%+v]", execution.Id, execution.Spec.LaunchPlan)
		return
	}

	if launchPlan.Spec.EntityMetadata == nil ||
		launchPlan.Spec.EntityMetadata.Schedule == nil ||
		launchPlan.Spec.EntityMetadata.Schedule.KickoffTimeInputArg == "" {
		// Kickoff time arguments aren't always required for scheduled workflows.
		logger.Debugf(context.Background(), "no kickoff time to report for scheduled workflow execution [%+v]",
			execution.Id)
		return
	}

	var inputs core.LiteralMap
	err = m.storageClient.ReadProtobuf(ctx, executionModel.InputsURI, &inputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to find inputs for emitting schedule delay event from uri: [%v]", executionModel.InputsURI)
		return
	}
	scheduledKickoffTimeProto := inputs.Literals[launchPlan.Spec.EntityMetadata.Schedule.KickoffTimeInputArg]
	if scheduledKickoffTimeProto == nil || scheduledKickoffTimeProto.GetScalar() == nil ||
		scheduledKickoffTimeProto.GetScalar().GetPrimitive() == nil ||
		scheduledKickoffTimeProto.GetScalar().GetPrimitive().GetDatetime() == nil {
		logger.Warningf(context.Background(),
			"failed to find scheduled kickoff time datetime value for scheduled workflow execution [%+v] "+
				"although one was expected", execution.Id)
		return
	}
	scheduledKickoffTime, err := ptypes.Timestamp(scheduledKickoffTimeProto.GetScalar().GetPrimitive().GetDatetime())
	if err != nil {
		// Timestamps are serialized by flyteadmin and should always be valid
		return
	}
	runningEventTime, err := ptypes.Timestamp(runningEventTimeProto)
	if err != nil {
		// Timestamps are always sent from propeller and should always be valid
		return
	}

	domainCounterMap, ok := m.userMetrics.ScheduledExecutionDelays[execution.Id.Project]
	if !ok {
		domainCounterMap = make(map[string]*promutils.StopWatch)
		m.userMetrics.ScheduledExecutionDelays[execution.Id.Project] = domainCounterMap
	}

	var watch *promutils.StopWatch
	watch, ok = domainCounterMap[execution.Id.Domain]
	if !ok {
		newWatch, err := m.systemMetrics.Scope.NewSubScope(execution.Id.Project).NewSubScope(execution.Id.Domain).NewStopWatch(
			"scheduled_execution_delay",
			"delay between scheduled execution time and time execution was observed running",
			time.Nanosecond)
		if err != nil {
			// Could be related to a concurrent exception.
			logger.Debugf(context.Background(),
				"failed to emit scheduled workflow execution delay stat, couldn't find or create counter")
			return
		}
		watch = &newWatch
		domainCounterMap[execution.Id.Domain] = watch
	}
	watch.Observe(scheduledKickoffTime, runningEventTime)
}

func (m *ExecutionManager) emitOverallWorkflowExecutionTime(
	executionModel *models.Execution, terminalEventTimeProto *timestamp.Timestamp) {
	if executionModel == nil || terminalEventTimeProto == nil {
		logger.Warningf(context.Background(),
			"tried to calculate scheduled workflow execution stats with a nil execution or event time")
		return
	}

	domainCounterMap, ok := m.userMetrics.WorkflowExecutionDurations[executionModel.Project]
	if !ok {
		domainCounterMap = make(map[string]*promutils.StopWatch)
		m.userMetrics.WorkflowExecutionDurations[executionModel.Project] = domainCounterMap
	}

	var watch *promutils.StopWatch
	watch, ok = domainCounterMap[executionModel.Domain]
	if !ok {
		newWatch, err := m.systemMetrics.Scope.NewSubScope(executionModel.Project).NewSubScope(executionModel.Domain).NewStopWatch(
			"workflow_execution_duration",
			"overall time from when when a workflow create request was sent to k8s to the workflow terminating",
			time.Nanosecond)
		if err != nil {
			// Could be related to a concurrent exception.
			logger.Debugf(context.Background(),
				"failed to emit workflow execution duration stat, couldn't find or create counter")
			return
		}
		watch = &newWatch
		domainCounterMap[executionModel.Domain] = watch
	}

	terminalEventTime, err := ptypes.Timestamp(terminalEventTimeProto)
	if err != nil {
		// Timestamps are always sent from propeller and should always be valid
		return
	}

	if executionModel.ExecutionCreatedAt == nil {
		logger.Warningf(context.Background(), "found execution with nil ExecutionCreatedAt: [%s/%s/%s]",
			executionModel.Project, executionModel.Domain, executionModel.Name)
		return
	}
	watch.Observe(*executionModel.ExecutionCreatedAt, terminalEventTime)
}

func (m *ExecutionManager) CreateWorkflowEvent(ctx context.Context, request admin.WorkflowExecutionEventRequest) (
	*admin.WorkflowExecutionEventResponse, error) {
	err := validation.ValidateCreateWorkflowEventRequest(request)
	if err != nil {
		logger.Debugf(ctx, "received invalid CreateWorkflowEventRequest [%s]: %v", request.RequestId, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Event.ExecutionId)
	logger.Debugf(ctx, "Received workflow execution event for [%+v] transitioning to phase [%v]",
		request.Event.ExecutionId, request.Event.Phase)

	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Event.ExecutionId)
	if err != nil {
		logger.Debugf(ctx, "failed to find execution [%+v] for recorded event [%s]: %v",
			request.Event.ExecutionId, request.RequestId, err)
		return nil, err
	}

	wfExecPhase := core.WorkflowExecution_Phase(core.WorkflowExecution_Phase_value[executionModel.Phase])
	if wfExecPhase == request.Event.Phase {
		logger.Debugf(ctx, "This phase %s was already recorded for workflow execution %v",
			wfExecPhase.String(), request.Event.ExecutionId)
		return nil, errors.NewFlyteAdminErrorf(codes.AlreadyExists,
			"This phase %s was already recorded for workflow execution %v",
			wfExecPhase.String(), request.Event.ExecutionId)
	} else if common.IsExecutionTerminal(wfExecPhase) {
		// Cannot go backwards in time from a terminal state to anything else
		curPhase := wfExecPhase.String()
		errorMsg := fmt.Sprintf("Invalid phase change from %s to %s for workflow execution %v", curPhase, request.Event.Phase.String(), request.Event.ExecutionId)
		return nil, errors.NewAlreadyInTerminalStateError(ctx, errorMsg, curPhase)
	}

	err = transformers.UpdateExecutionModelState(executionModel, request)
	if err != nil {
		logger.Debugf(ctx, "failed to transform updated workflow execution model [%+v] after receiving event with err: %v",
			request.Event.ExecutionId, err)
		return nil, err
	}
	err = m.db.ExecutionRepo().Update(ctx, *executionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to update execution with CreateWorkflowEvent [%+v] with err %v",
			request, err)
		return nil, err
	}
	m.dbEventWriter.Write(request)

	if request.Event.Phase == core.WorkflowExecution_RUNNING {
		// Workflow executions are created in state "UNDEFINED". All the time up until a RUNNING event is received is
		// considered system-induced delay.
		if executionModel.Mode == int32(admin.ExecutionMetadata_SCHEDULED) {
			go m.emitScheduledWorkflowMetrics(ctx, executionModel, request.Event.OccurredAt)
		}
	} else if common.IsExecutionTerminal(request.Event.Phase) {
		m.systemMetrics.ActiveExecutions.Dec()
		m.systemMetrics.ExecutionsTerminated.Inc()
		go m.emitOverallWorkflowExecutionTime(executionModel, request.Event.OccurredAt)

		err = m.publishNotifications(ctx, request, *executionModel)
		if err != nil {
			// The only errors that publishNotifications will forward are those related
			// to unexpected data and transformation errors.
			logger.Debugf(ctx, "failed to publish notifications for CreateWorkflowEvent [%+v] due to err: %v",
				request, err)
			return nil, err
		}
	}
	if err := m.eventPublisher.Publish(ctx, proto.MessageName(&request), &request); err != nil {
		m.systemMetrics.PublishEventError.Inc()
		logger.Infof(ctx, "error publishing event [%+v] with err: [%v]", request.RequestId, err)
	}

	m.systemMetrics.ExecutionEventsCreated.Inc()
	return &admin.WorkflowExecutionEventResponse{}, nil
}

func (m *ExecutionManager) GetExecution(
	ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "GetExecution request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Id)
	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}
	execution, transformerErr := transformers.FromExecutionModel(*executionModel)
	if transformerErr != nil {
		logger.Debugf(ctx, "Failed to transform execution model [%+v] to proto object with err: %v", request.Id,
			transformerErr)
		return nil, transformerErr
	}

	return execution, nil
}

func (m *ExecutionManager) GetExecutionData(
	ctx context.Context, request admin.WorkflowExecutionGetDataRequest) (*admin.WorkflowExecutionGetDataResponse, error) {
	ctx = getExecutionContext(ctx, request.Id)
	executionModel, err := util.GetExecutionModel(ctx, m.db, *request.Id)
	if err != nil {
		logger.Debugf(ctx, "Failed to get execution model for request [%+v] with err: %v", request, err)
		return nil, err
	}
	execution, err := transformers.FromExecutionModel(*executionModel)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform execution model [%+v] to proto object with err: %v", request.Id, err)
		return nil, err
	}
	signedOutputsURLBlob := admin.UrlBlob{}
	if execution.Closure.GetOutputs() != nil && execution.Closure.GetOutputs().GetUri() != "" {
		signedOutputsURLBlob, err = m.urlData.Get(ctx, execution.Closure.GetOutputs().GetUri())
		if err != nil {
			return nil, err
		}
	}
	// Prior to flyteidl v0.15.0, Inputs were held in ExecutionClosure and were not offloaded. Ensure we can return the inputs as expected.
	if len(executionModel.InputsURI) == 0 {
		closure := &admin.ExecutionClosure{}
		// We must not use the FromExecutionModel method because it empties deprecated fields.
		if err := proto.Unmarshal(executionModel.Closure, closure); err != nil {
			return nil, err
		}
		newInputsURI, err := m.offloadInputs(ctx, closure.ComputedInputs, request.Id, shared.Inputs)
		if err != nil {
			return nil, err
		}
		// Update model so as not to offload again.
		executionModel.InputsURI = newInputsURI
		if err := m.db.ExecutionRepo().Update(ctx, *executionModel); err != nil {
			return nil, err
		}
	}
	inputsURLBlob, err := m.urlData.Get(ctx, executionModel.InputsURI.String())
	if err != nil {
		return nil, err
	}
	response := &admin.WorkflowExecutionGetDataResponse{
		Outputs: &signedOutputsURLBlob,
		Inputs:  &inputsURLBlob,
	}
	maxDataSize := m.config.ApplicationConfiguration().GetRemoteDataConfig().MaxSizeInBytes
	remoteDataScheme := m.config.ApplicationConfiguration().GetRemoteDataConfig().Scheme
	if util.ShouldFetchData(m.config.ApplicationConfiguration().GetRemoteDataConfig(), inputsURLBlob) {
		var fullInputs core.LiteralMap
		err := m.storageClient.ReadProtobuf(ctx, executionModel.InputsURI, &fullInputs)
		if err != nil {
			logger.Warningf(ctx, "Failed to read inputs from URI [%s] with err: %v", executionModel.InputsURI, err)
		}
		response.FullInputs = &fullInputs
	}
	if remoteDataScheme == common.Local || remoteDataScheme == common.None || (signedOutputsURLBlob.Bytes < maxDataSize && execution.Closure.GetOutputs() != nil) {
		var fullOutputs core.LiteralMap
		outputsURI := execution.Closure.GetOutputs().GetUri()
		err := m.storageClient.ReadProtobuf(ctx, storage.DataReference(outputsURI), &fullOutputs)
		if err != nil {
			logger.Warningf(ctx, "Failed to read outputs from URI [%s] with err: %v", outputsURI, err)
		}
		response.FullOutputs = &fullOutputs
	}

	m.userMetrics.WorkflowExecutionInputBytes.Observe(float64(response.Inputs.Bytes))
	m.userMetrics.WorkflowExecutionOutputBytes.Observe(float64(response.Outputs.Bytes))
	return response, nil
}

func (m *ExecutionManager) ListExecutions(
	ctx context.Context, request admin.ResourceListRequest) (*admin.ExecutionList, error) {
	// Check required fields
	if err := validation.ValidateResourceListRequest(request); err != nil {
		logger.Debugf(ctx, "ListExecutions request [%+v] failed validation with err: %v", request, err)
		return nil, err
	}
	ctx = contextutils.WithProjectDomain(ctx, request.Id.Project, request.Id.Domain)
	filters, err := util.GetDbFilters(util.FilterSpec{
		Project:        request.Id.Project,
		Domain:         request.Id.Domain,
		Name:           request.Id.Name, // Optional, may be empty.
		RequestFilters: request.Filters,
	}, common.Execution)
	if err != nil {
		return nil, err
	}
	var sortParameter common.SortParameter
	if request.SortBy != nil {
		sortParameter, err = common.NewSortParameter(*request.SortBy)
		if err != nil {
			return nil, err
		}
	}

	offset, err := validation.ValidateToken(request.Token)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid pagination token %s for ListExecutions",
			request.Token)
	}
	joinTableEntities := make(map[common.Entity]bool)
	for _, filter := range filters {
		joinTableEntities[filter.GetEntity()] = true
	}
	listExecutionsInput := repositoryInterfaces.ListResourceInput{
		Limit:             int(request.Limit),
		Offset:            offset,
		InlineFilters:     filters,
		SortParameter:     sortParameter,
		JoinTableEntities: joinTableEntities,
	}
	output, err := m.db.ExecutionRepo().List(ctx, listExecutionsInput)
	if err != nil {
		logger.Debugf(ctx, "Failed to list executions using input [%+v] with err %v", listExecutionsInput, err)
		return nil, err
	}
	executionList, err := transformers.FromExecutionModels(output.Executions)
	if err != nil {
		logger.Errorf(ctx,
			"Failed to transform execution models [%+v] with err: %v", output.Executions, err)
		return nil, err
	}
	// TODO: TO BE DELETED
	// Clear deprecated fields during migration phase. Once migration is complete, these will be cleared in the database.
	// Thus this will be redundant
	for _, execution := range executionList {
		execution.Spec.Inputs = nil
		execution.Closure.ComputedInputs = nil
	}
	// END TO BE DELETED
	var token string
	if len(executionList) == int(request.Limit) {
		token = strconv.Itoa(offset + len(executionList))
	}
	return &admin.ExecutionList{
		Executions: executionList,
		Token:      token,
	}, nil
}

// publishNotifications will only forward major errors because the assumption made is all of the objects
// that are being manipulated have already been validated/manipulated by Flyte itself.
// Note: This method should be refactored somewhere else once the interaction with pushing to SNS.
func (m *ExecutionManager) publishNotifications(ctx context.Context, request admin.WorkflowExecutionEventRequest,
	execution models.Execution) error {
	// Notifications are stored in the Spec object of an admin.Execution object.
	adminExecution, err := transformers.FromExecutionModel(execution)
	if err != nil {
		// This shouldn't happen because execution manager marshaled the data into models.Execution.
		m.systemMetrics.TransformerError.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "Failed to transform execution [%+v] with err: %v", request.Event.ExecutionId, err)
	}
	var notificationsList = adminExecution.Closure.Notifications
	logger.Debugf(ctx, "publishing notifications for execution [%+v] in state [%+v] for notifications [%+v]",
		request.Event.ExecutionId, request.Event.Phase, notificationsList)
	for _, notification := range notificationsList {
		// Check if the notification phase matches the current one.
		var matchPhase = false
		for _, phase := range notification.Phases {
			if phase == request.Event.Phase {
				matchPhase = true
			}
		}

		// The current phase doesn't match; no notifications will be sent for the current notification option.
		if !matchPhase {
			continue
		}

		// Currently all three supported notifications use email underneath to send the notification.
		// Convert Slack and PagerDuty into an EmailNotification type.
		var emailNotification admin.EmailNotification
		if notification.GetEmail() != nil {
			emailNotification.RecipientsEmail = notification.GetEmail().GetRecipientsEmail()
		} else if notification.GetPagerDuty() != nil {
			emailNotification.RecipientsEmail = notification.GetPagerDuty().GetRecipientsEmail()
		} else if notification.GetSlack() != nil {
			emailNotification.RecipientsEmail = notification.GetSlack().GetRecipientsEmail()
		} else {
			logger.Debugf(ctx, "failed to publish notification, encountered unrecognized type: %v", notification.Type)
			m.systemMetrics.UnexpectedDataError.Inc()
			// Unsupported notification types should have been caught when the launch plan was being created.
			return errors.NewFlyteAdminErrorf(codes.Internal, "Unsupported notification type [%v] for execution [%+v]",
				notification.Type, request.Event.ExecutionId)
		}

		// Convert the email Notification into an email message to be published.
		// Currently there are no possible errors while creating an email message.
		// Once customizable content is specified, errors are possible.
		email := notifications.ToEmailMessageFromWorkflowExecutionEvent(
			*m.config.ApplicationConfiguration().GetNotificationsConfig(), emailNotification, request, adminExecution)
		// Errors seen while publishing a message are considered non-fatal to the method and will not result
		// in the method returning an error.
		if err = m.notificationClient.Publish(ctx, proto.MessageName(&emailNotification), email); err != nil {
			m.systemMetrics.PublishNotificationError.Inc()
			logger.Infof(ctx, "error publishing email notification [%+v] with err: [%v]", notification, err)
		}
	}
	return nil
}

func (m *ExecutionManager) TerminateExecution(
	ctx context.Context, request admin.ExecutionTerminateRequest) (*admin.ExecutionTerminateResponse, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(request.Id); err != nil {
		logger.Debugf(ctx, "received terminate execution request: %v with invalid identifier: %v", request, err)
		return nil, err
	}
	ctx = getExecutionContext(ctx, request.Id)
	// Save the abort reason (best effort)
	executionModel, err := m.db.ExecutionRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: request.Id.Project,
		Domain:  request.Id.Domain,
		Name:    request.Id.Name,
	})
	if err != nil {
		logger.Infof(ctx, "couldn't find execution [%+v] to save termination cause", request.Id)
		return nil, err
	}

	err = m.workflowExecutor.TerminateWorkflowExecution(ctx, workflowengineInterfaces.TerminateWorkflowInput{
		ExecutionID: request.Id,
		Cluster:     executionModel.Cluster,
	})
	if err != nil {
		return nil, err
	}

	err = transformers.SetExecutionAborted(&executionModel, request.Cause, getUser(ctx))
	if err != nil {
		logger.Debugf(ctx, "failed to add abort metadata for execution [%+v] with err: %v", request.Id, err)
		return nil, err
	}
	err = m.db.ExecutionRepo().Update(ctx, executionModel)
	if err != nil {
		logger.Debugf(ctx, "failed to save abort cause for terminated execution: %+v with err: %v", request.Id, err)
		return nil, err
	}
	return &admin.ExecutionTerminateResponse{}, nil
}

func newExecutionSystemMetrics(scope promutils.Scope) executionSystemMetrics {
	return executionSystemMetrics{
		Scope: scope,
		ActiveExecutions: scope.MustNewGauge("active_executions",
			"overall count of active workflow executions"),
		ExecutionsCreated: scope.MustNewCounter("executions_created",
			"overall count of successfully completed CreateExecutionRequests"),
		ExecutionsTerminated: scope.MustNewCounter("executions_terminated",
			"overall count of terminated workflow executions"),
		ExecutionEventsCreated: scope.MustNewCounter("execution_events_created",
			"overall count of successfully completed WorkflowExecutionEventRequest"),
		PropellerFailures: scope.MustNewCounter("propeller_failures",
			"propeller failures in creating workflow executions"),
		TransformerError: scope.MustNewCounter("transformer_error",
			"overall count of errors when transforming models and messages"),
		UnexpectedDataError: scope.MustNewCounter("unexpected_data_error",
			"overall count of unexpected data for previously validated objects"),
		PublishNotificationError: scope.MustNewCounter("publish_error",
			"overall count of publish notification errors when invoking publish()"),
		SpecSizeBytes:    scope.MustNewSummary("spec_size_bytes", "size in bytes of serialized execution spec"),
		ClosureSizeBytes: scope.MustNewSummary("closure_size_bytes", "size in bytes of serialized execution closure"),
		AcceptanceDelay: scope.MustNewSummary("acceptance_delay",
			"delay in seconds from when an execution was requested to be created and when it actually was"),
		PublishEventError: scope.MustNewCounter("publish_event_error",
			"overall count of publish event errors when invoking publish()"),
	}
}

func NewExecutionManager(db repositories.RepositoryInterface, config runtimeInterfaces.Configuration,
	storageClient *storage.DataStore, workflowExecutor workflowengineInterfaces.Executor, systemScope promutils.Scope,
	userScope promutils.Scope, publisher notificationInterfaces.Publisher, urlData dataInterfaces.RemoteURLInterface,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface,
	eventPublisher notificationInterfaces.Publisher, eventWriter eventWriter.WorkflowExecutionEventWriter) interfaces.ExecutionInterface {
	queueAllocator := executions.NewQueueAllocator(config, db)
	systemMetrics := newExecutionSystemMetrics(systemScope)

	userMetrics := executionUserMetrics{
		Scope:                      userScope,
		ScheduledExecutionDelays:   make(map[string]map[string]*promutils.StopWatch),
		WorkflowExecutionDurations: make(map[string]map[string]*promutils.StopWatch),
		WorkflowExecutionInputBytes: userScope.MustNewSummary("input_size_bytes",
			"size in bytes of serialized execution inputs"),
		WorkflowExecutionOutputBytes: userScope.MustNewSummary("output_size_bytes",
			"size in bytes of serialized execution outputs"),
	}

	resourceManager := resources.NewResourceManager(db, config.ApplicationConfiguration())
	return &ExecutionManager{
		db:                        db,
		config:                    config,
		storageClient:             storageClient,
		workflowExecutor:          workflowExecutor,
		queueAllocator:            queueAllocator,
		_clock:                    clock.New(),
		systemMetrics:             systemMetrics,
		userMetrics:               userMetrics,
		notificationClient:        publisher,
		urlData:                   urlData,
		workflowManager:           workflowManager,
		namedEntityManager:        namedEntityManager,
		resourceManager:           resourceManager,
		qualityOfServiceAllocator: executions.NewQualityOfServiceAllocator(config, resourceManager),
		eventPublisher:            eventPublisher,
		dbEventWriter:             eventWriter,
	}
}

// Adds project labels with higher precedence to workflow labels. Project labels are ignored if a corresponding label is set on the workflow.
func (m *ExecutionManager) addProjectLabels(ctx context.Context, projectName string, initialLabels map[string]string) (map[string]string, error) {
	project, err := m.db.ProjectRepo().Get(ctx, projectName)
	if err != nil {
		logger.Errorf(ctx, "Failed to get project for [%+v] with error: %v", project, err)
		return nil, err
	}
	// passing nil domain as not needed to retrieve labels
	projectLabels := transformers.FromProjectModel(project, nil).Labels.GetValues()

	if initialLabels == nil {
		initialLabels = make(map[string]string)
	}

	for k, v := range projectLabels {
		if _, ok := initialLabels[k]; !ok {
			initialLabels[k] = v
		}
	}
	return initialLabels, nil
}
